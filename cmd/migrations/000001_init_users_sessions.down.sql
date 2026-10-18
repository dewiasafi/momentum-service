DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token_hash;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;