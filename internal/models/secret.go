package models

import (
	"time"
)

// Secret представляет секрет в базе данных
type Secret struct {
	ID               string     `json:"id" db:"id"`
	EncryptedContent string     `json:"-" db:"encrypted_content"`         // Зашифрованный контент (не отдаём в JSON)
	IV               string     `json:"-" db:"iv"`                        // Initialization vector для расшифровки
	ExpiresAt        time.Time  `json:"expires_at" db:"expires_at"`       // Время истечения срока действия
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`       // Время создания
	AccessedAt       *time.Time `json:"accessed_at,omitempty" db:"accessed_at"` // Время первого доступа (nullable)
	IsAccessed       bool       `json:"is_accessed" db:"is_accessed"`     // Был ли прочитан
}

// IsExpired проверяет, истёк ли срок действия секрета
func (s *Secret) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid проверяет, можно ли прочитать секрет
func (s *Secret) IsValid() bool {
	return !s.IsExpired() && !s.IsAccessed
}

// CreateSecretRequest представляет запрос на создание секрета
type CreateSecretRequest struct {
	Content        string `json:"content" binding:"required"`          // Текст секрета
	ExpirationHours int    `json:"expiration_hours" binding:"required,oneof=24 48 72"` // Время жизни: 24, 48 или 72 часа
}

// CreateSecretResponse представляет ответ после создания секрета
type CreateSecretResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetSecretResponse представляет ответ при получении секрета
type GetSecretResponse struct {
	Content   string    `json:"content"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
