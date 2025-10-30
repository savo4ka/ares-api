package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/savo4ka/ares-api/internal/database"
	"github.com/savo4ka/ares-api/internal/models"
)

// SecretRepository предоставляет методы для работы с секретами в БД
type SecretRepository struct {
	db *database.DB
}

// NewSecretRepository создаёт новый репозиторий секретов
func NewSecretRepository(db *database.DB) *SecretRepository {
	return &SecretRepository{
		db: db,
	}
}

// Create создаёт новый секрет в базе данных
func (r *SecretRepository) Create(secret *models.Secret) error {
	query := `
		INSERT INTO secrets (id, encrypted_content, iv, expires_at, created_at, is_accessed)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		query,
		secret.ID,
		secret.EncryptedContent,
		secret.IV,
		secret.ExpiresAt,
		secret.CreatedAt,
		secret.IsAccessed,
	)

	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// GetByID получает секрет по ID
func (r *SecretRepository) GetByID(id string) (*models.Secret, error) {
	query := `
		SELECT id, encrypted_content, iv, expires_at, created_at, accessed_at, is_accessed
		FROM secrets
		WHERE id = $1
	`

	secret := &models.Secret{}
	err := r.db.QueryRow(query, id).Scan(
		&secret.ID,
		&secret.EncryptedContent,
		&secret.IV,
		&secret.ExpiresAt,
		&secret.CreatedAt,
		&secret.AccessedAt,
		&secret.IsAccessed,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("secret not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return secret, nil
}

// MarkAsAccessed помечает секрет как прочитанный
func (r *SecretRepository) MarkAsAccessed(id string) error {
	query := `
		UPDATE secrets
		SET is_accessed = TRUE, accessed_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark secret as accessed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("secret not found")
	}

	return nil
}

// CleanupExpired удаляет истёкшие секреты из базы данных
func (r *SecretRepository) CleanupExpired() (int64, error) {
	query := `
		DELETE FROM secrets
		WHERE expires_at < $1
	`

	result, err := r.db.Exec(query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired secrets: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// DeleteByID удаляет секрет по ID (для тестирования или админских функций)
func (r *SecretRepository) DeleteByID(id string) error {
	query := `DELETE FROM secrets WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("secret not found")
	}

	return nil
}
