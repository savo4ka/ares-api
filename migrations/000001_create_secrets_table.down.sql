-- Удаление таблицы при откате миграции
DROP INDEX IF EXISTS idx_secrets_is_accessed;
DROP INDEX IF EXISTS idx_secrets_expires_at;
DROP TABLE IF EXISTS secrets;
