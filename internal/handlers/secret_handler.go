package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/savo4ka/ares-api/internal/crypto"
	"github.com/savo4ka/ares-api/internal/models"
	"github.com/savo4ka/ares-api/internal/repository"
)

// SecretHandler обрабатывает HTTP запросы для работы с секретами
type SecretHandler struct {
	repo              *repository.SecretRepository
	encryptionService *crypto.EncryptionService
	baseURL           string
}

// NewSecretHandler создаёт новый обработчик секретов
func NewSecretHandler(repo *repository.SecretRepository, encryptionService *crypto.EncryptionService, baseURL string) *SecretHandler {
	return &SecretHandler{
		repo:              repo,
		encryptionService: encryptionService,
		baseURL:           baseURL,
	}
}

// CreateSecret обрабатывает POST /api/secrets - создание нового секрета
func (h *SecretHandler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSecretRequest

	// Парсим JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Валидация
	if req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Content is required")
		return
	}

	if req.ExpirationHours != 24 && req.ExpirationHours != 48 && req.ExpirationHours != 72 {
		respondWithError(w, http.StatusBadRequest, "Expiration hours must be 24, 48, or 72")
		return
	}

	// Шифруем контент
	encryptedData, err := h.encryptionService.Encrypt(req.Content)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to encrypt secret")
		return
	}

	// Создаём модель секрета
	secret := &models.Secret{
		ID:               uuid.New().String(),
		EncryptedContent: encryptedData.Ciphertext,
		IV:               encryptedData.IV,
		ExpiresAt:        time.Now().Add(time.Duration(req.ExpirationHours) * time.Hour),
		CreatedAt:        time.Now(),
		IsAccessed:       false,
	}

	// Сохраняем в базу данных
	if err := h.repo.Create(secret); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create secret")
		return
	}

	// Формируем URL для доступа к секрету
	secretURL := fmt.Sprintf("%s/secret/%s", h.baseURL, secret.ID)

	// Возвращаем ответ
	response := models.CreateSecretResponse{
		ID:        secret.ID,
		URL:       secretURL,
		ExpiresAt: secret.ExpiresAt,
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// GetSecret обрабатывает GET /api/secrets/:id - получение секрета
func (h *SecretHandler) GetSecret(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Secret ID is required")
		return
	}

	// Получаем секрет из БД
	secret, err := h.repo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Secret not found")
		return
	}

	// Проверяем, не истёк ли срок действия
	if secret.IsExpired() {
		respondWithError(w, http.StatusGone, "Secret has expired")
		return
	}

	// Проверяем, не был ли уже прочитан
	if secret.IsAccessed {
		respondWithError(w, http.StatusGone, "Secret has already been accessed")
		return
	}

	// Расшифровываем контент
	encryptedData := &crypto.EncryptedData{
		Ciphertext: secret.EncryptedContent,
		IV:         secret.IV,
	}

	plaintext, err := h.encryptionService.Decrypt(encryptedData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decrypt secret")
		return
	}

	// Помечаем секрет как прочитанный
	if err := h.repo.MarkAsAccessed(id); err != nil {
		// Логируем ошибку, но не возвращаем пользователю, т.к. секрет уже расшифрован
		fmt.Printf("Warning: failed to mark secret as accessed: %v\n", err)
	}

	// Возвращаем расшифрованный контент
	response := models.GetSecretResponse{
		Content:   plaintext,
		ExpiresAt: secret.ExpiresAt,
		CreatedAt: secret.CreatedAt,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HealthCheck обрабатывает GET /health - проверка здоровья сервиса
func (h *SecretHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// respondWithError отправляет JSON ошибку
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON отправляет JSON ответ
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
