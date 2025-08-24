-- Rollback users table creation
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP TABLE IF EXISTS users;