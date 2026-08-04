-- RealTimeMessageChat Database Schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    user_id     VARCHAR(255) UNIQUE NOT NULL,
    username    VARCHAR(255),
    email       VARCHAR(255),
    status      VARCHAR(20) NOT NULL DEFAULT 'offline',
    last_seen   TIMESTAMP WITH TIME ZONE,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_user_id ON users(user_id);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

COMMENT ON TABLE users IS 'API users managed through /api/v1/users';

CREATE TABLE IF NOT EXISTS messages (
    id              BIGSERIAL PRIMARY KEY,

    message_id      UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),

    user_id         VARCHAR(255) NOT NULL,
    sub_id          VARCHAR(255) NOT NULL DEFAULT '',
    command         VARCHAR(255) NOT NULL DEFAULT '',

    publisher_info  TEXT NOT NULL DEFAULT '{}',
    server_name     VARCHAR(100) NOT NULL DEFAULT 'MainServer',

    content         TEXT NOT NULL,
    is_encrypted    BOOLEAN NOT NULL DEFAULT FALSE,

    status          VARCHAR(20) NOT NULL DEFAULT 'pending',

    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_messages_message_id ON messages(message_id);
CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);
CREATE INDEX IF NOT EXISTS idx_messages_sub_id ON messages(sub_id);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_server_name ON messages(server_name);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_user_id_created_at ON messages(user_id, created_at DESC);

COMMENT ON TABLE messages IS 'Broadcast messages consumed from RabbitMQ, optionally encrypted';
COMMENT ON COLUMN messages.message_id IS 'Producer-supplied tracking UUID (publisher_information.message_id)';
COMMENT ON COLUMN messages.user_id IS 'User identifier from the message JSON "id" field';
COMMENT ON COLUMN messages.sub_id IS 'Session identifier from the message JSON "sub_id" field';
COMMENT ON COLUMN messages.command IS 'Command from the message JSON "message.command" field';
COMMENT ON COLUMN messages.publisher_info IS 'JSON string of producer metadata';
COMMENT ON COLUMN messages.content IS 'Message body - encrypted (base64) or plain text';
COMMENT ON COLUMN messages.is_encrypted IS 'TRUE if content is encrypted';
COMMENT ON COLUMN messages.status IS 'pending, sent, processed, or failed';

CREATE OR REPLACE VIEW recent_messages AS
SELECT
    id,
    message_id,
    user_id,
    sub_id,
    command,
    server_name,
    is_encrypted,
    status,
    created_at,
    CASE
        WHEN is_encrypted THEN '[ENCRYPTED]'
        ELSE LEFT(content, 100)
    END AS content_preview
FROM messages
WHERE created_at >= NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;

COMMENT ON VIEW recent_messages IS 'Messages from the last 24 hours with content preview';

CREATE OR REPLACE VIEW message_statistics AS
SELECT
    DATE(created_at) AS message_date,
    server_name,
    is_encrypted,
    COUNT(*) AS message_count,
    COUNT(DISTINCT user_id) AS unique_users,
    MIN(created_at) AS first_message,
    MAX(created_at) AS last_message
FROM messages
GROUP BY DATE(created_at), server_name, is_encrypted
ORDER BY message_date DESC, server_name;

COMMENT ON VIEW message_statistics IS 'Daily message statistics by server and encryption status';

CREATE OR REPLACE FUNCTION cleanup_old_messages(days_to_keep INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM messages
    WHERE created_at < NOW() - (days_to_keep || ' days')::INTERVAL;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_messages IS 'Delete messages older than the given number of days (default 30). 호출 스케줄러는 없다 - 수동 실행 전제.';

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS encryption_keys (
    key_id          BIGSERIAL PRIMARY KEY,
    key_name        VARCHAR(100) UNIQUE NOT NULL,
    encryption_key  TEXT NOT NULL,
    encryption_iv   TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    rotated_at      TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_encryption_keys_active ON encryption_keys(is_active) WHERE is_active = TRUE;

COMMENT ON TABLE encryption_keys IS 'AES-256-CBC key storage. NOT YET USED - keys come from the consumer config file.';
