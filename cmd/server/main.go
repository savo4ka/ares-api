package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/savo4ka/ares-api/internal/config"
	"github.com/savo4ka/ares-api/internal/crypto"
	"github.com/savo4ka/ares-api/internal/database"
	"github.com/savo4ka/ares-api/internal/handlers"
	"github.com/savo4ka/ares-api/internal/metrics"
	"github.com/savo4ka/ares-api/internal/repository"
)

func main() {
	// Загружаем .env файл (для локальной разработки)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к базе данных
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	// Инициализируем сервис шифрования
	encryptionService, err := crypto.NewEncryptionService(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("Failed to create encryption service: %v", err)
	}

	// Инициализируем метрики
	appMetrics := metrics.New()

	// Создаём репозиторий
	secretRepo := repository.NewSecretRepository(db)

	// Создаём handlers
	baseURL := fmt.Sprintf("http://localhost:%s", cfg.ServerPort)
	secretHandler := handlers.NewSecretHandler(secretRepo, encryptionService, baseURL, appMetrics)

	// Настраиваем роутер
	router := mux.NewRouter()

	// Применяем middleware
	router.Use(handlers.CORSMiddleware(cfg.AllowedOrigins))
	router.Use(handlers.LoggingMiddleware)
	router.Use(handlers.MetricsMiddleware(appMetrics))

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/secrets", secretHandler.CreateSecret).Methods("POST", "OPTIONS")
	api.HandleFunc("/secrets/{id}", secretHandler.GetSecret).Methods("GET", "OPTIONS")

	// Health check
	router.HandleFunc("/health", secretHandler.HealthCheck).Methods("GET")

	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Запускаем фоновую задачу по очистке истёкших секретов
	go cleanupExpiredSecrets(secretRepo, appMetrics)

	// Настраиваем HTTP сервер
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Канал для обработки системных сигналов
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server is starting on port %s...", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("Server started successfully")

	// Ждём сигнала о завершении
	<-done
	log.Println("Server is shutting down...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

// cleanupExpiredSecrets периодически удаляет истёкшие секреты из БД
func cleanupExpiredSecrets(repo *repository.SecretRepository, m *metrics.Metrics) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		deleted, err := repo.CleanupExpired()
		if err != nil {
			log.Printf("Failed to cleanup expired secrets: %v", err)
			continue
		}

		if deleted > 0 {
			// Добавляем количество удалённых секретов к метрике
			m.SecretsCleanedUpTotal.Add(float64(deleted))
			log.Printf("Cleaned up %d expired secrets", deleted)
		}

		// Обновляем метрику активных секретов
		if count, err := repo.GetActiveSecretsCount(); err == nil {
			m.UpdateActiveSecretsGauge(count)
		}
	}
}
