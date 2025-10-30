-- +migrate Up
-- Создание таблицы для хранения секретов
CREATE TABLE IF NOT EXISTS secrets (
    id VARCHAR(36) PRIMARY KEY,
    encrypted_content TEXT NOT NULL,
    iv VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    accessed_at TIMESTAMP WITH TIME ZONE,
    is_accessed BOOLEAN NOT NULL DEFAULT FALSE
);

-- Создание индекса для поиска по времени истечения (для очистки старых записей)
CREATE INDEX IF NOT EXISTS idx_secrets_expires_at ON secrets(expires_at);

-- Создание индекса для поиска по is_accessed
CREATE INDEX IF NOT EXISTS idx_secrets_is_accessed ON secrets(is_accessed);

-- +migrate Down
-- Удаление таблицы при откате миграции
DROP INDEX IF EXISTS idx_secrets_is_accessed;
DROP INDEX IF EXISTS idx_secrets_expires_at;
DROP TABLE IF EXISTS secrets;
