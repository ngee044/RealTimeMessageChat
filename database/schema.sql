-- RealTimeMessageChat Database Schema
-- PostgreSQL database schema for storing encrypted broadcast messages

-- Create database (run as postgres superuser)
-- CREATE DATABASE rtmc;
-- CREATE USER rtmc_user WITH PASSWORD 'rtmc_password';
-- GRANT ALL PRIVILEGES ON DATABASE rtmc TO rtmc_user;

-- Connect to rtmc database before running the rest of this script
-- \c rtmc

-- Enable UUID extension for generating unique identifiers
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Messages table: stores all broadcast messages consumed from RabbitMQ
CREATE TABLE IF NOT EXISTS messages (
    -- Primary key: auto-incrementing sequence
    message_id BIGSERIAL PRIMARY KEY,

    -- User identification (from message JSON)
    id VARCHAR(255) NOT NULL,
    sub_id VARCHAR(255) NOT NULL,

    -- Publisher metadata (JSON string)
    publisher_info TEXT,

    -- Target server information
    server_name VARCHAR(100) NOT NULL DEFAULT 'MainServer',

    -- Message content (encrypted or plain text)
    message_content TEXT NOT NULL,

    -- Encryption flag
    is_encrypted BOOLEAN NOT NULL DEFAULT FALSE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Indexing for common queries
    CONSTRAINT messages_id_sub_id_idx UNIQUE (id, sub_id, created_at)
);

-- Indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_messages_id ON messages(id);
CREATE INDEX IF NOT EXISTS idx_messages_sub_id ON messages(sub_id);
CREATE INDEX IF NOT EXISTS idx_messages_server_name ON messages(server_name);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_is_encrypted ON messages(is_encrypted);

-- Composite index for common query patterns
CREATE INDEX IF NOT EXISTS idx_messages_id_created_at ON messages(id, created_at DESC);

-- Comments for documentation
COMMENT ON TABLE messages IS 'Stores broadcast messages consumed from RabbitMQ, optionally encrypted';
COMMENT ON COLUMN messages.message_id IS 'Auto-incrementing primary key';
COMMENT ON COLUMN messages.id IS 'User identifier from message JSON';
COMMENT ON COLUMN messages.sub_id IS 'Session/subscription identifier from message JSON';
COMMENT ON COLUMN messages.publisher_info IS 'JSON string containing publisher metadata';
COMMENT ON COLUMN messages.server_name IS 'Target server name (e.g., MainServer)';
COMMENT ON COLUMN messages.message_content IS 'Message content - encrypted (base64) or plain text';
COMMENT ON COLUMN messages.is_encrypted IS 'TRUE if message_content is encrypted, FALSE otherwise';
COMMENT ON COLUMN messages.created_at IS 'Timestamp when message was stored';

-- Grant permissions to rtmc_user
GRANT ALL PRIVILEGES ON TABLE messages TO rtmc_user;
GRANT USAGE, SELECT ON SEQUENCE messages_message_id_seq TO rtmc_user;

-- Optional: Create a view for recent messages (last 24 hours)
CREATE OR REPLACE VIEW recent_messages AS
SELECT
    message_id,
    id,
    sub_id,
    server_name,
    is_encrypted,
    created_at,
    CASE
        WHEN is_encrypted THEN '[ENCRYPTED]'
        ELSE LEFT(message_content, 100)
    END AS content_preview
FROM messages
WHERE created_at >= NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;

COMMENT ON VIEW recent_messages IS 'View showing messages from the last 24 hours with content preview';

GRANT SELECT ON recent_messages TO rtmc_user;

-- Optional: Function to clean up old messages (for maintenance)
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

COMMENT ON FUNCTION cleanup_old_messages IS 'Delete messages older than specified days (default: 30)';

GRANT EXECUTE ON FUNCTION cleanup_old_messages TO rtmc_user;

-- Optional: Create a table for encryption keys management
CREATE TABLE IF NOT EXISTS encryption_keys (
    key_id SERIAL PRIMARY KEY,
    key_name VARCHAR(100) UNIQUE NOT NULL,
    encryption_key TEXT NOT NULL,
    encryption_iv TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    rotated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_encryption_keys_active ON encryption_keys(is_active) WHERE is_active = TRUE;

COMMENT ON TABLE encryption_keys IS 'Stores encryption keys for message encryption (AES-256-CBC)';
COMMENT ON COLUMN encryption_keys.key_name IS 'Unique identifier for the key';
COMMENT ON COLUMN encryption_keys.encryption_key IS 'Base64-encoded AES encryption key';
COMMENT ON COLUMN encryption_keys.encryption_iv IS 'Base64-encoded initialization vector';
COMMENT ON COLUMN encryption_keys.is_active IS 'TRUE if this key is currently in use';
COMMENT ON COLUMN encryption_keys.rotated_at IS 'Timestamp when key was rotated/deactivated';

GRANT ALL PRIVILEGES ON TABLE encryption_keys TO rtmc_user;
GRANT USAGE, SELECT ON SEQUENCE encryption_keys_key_id_seq TO rtmc_user;

-- Insert a sample active encryption key (replace with actual key in production)
-- Note: Generate actual keys using Encryptor::create_key() from the C++ code
INSERT INTO encryption_keys (key_name, encryption_key, encryption_iv, is_active)
VALUES (
    'default_key',
    'REPLACE_WITH_ACTUAL_BASE64_KEY',
    'REPLACE_WITH_ACTUAL_BASE64_IV',
    FALSE  -- Set to TRUE after replacing with actual keys
) ON CONFLICT (key_name) DO NOTHING;

-- Optional: Create statistics view
CREATE OR REPLACE VIEW message_statistics AS
SELECT
    DATE(created_at) AS message_date,
    server_name,
    is_encrypted,
    COUNT(*) AS message_count,
    COUNT(DISTINCT id) AS unique_users,
    MIN(created_at) AS first_message,
    MAX(created_at) AS last_message
FROM messages
GROUP BY DATE(created_at), server_name, is_encrypted
ORDER BY message_date DESC, server_name;

COMMENT ON VIEW message_statistics IS 'Daily statistics of messages by server and encryption status';

GRANT SELECT ON message_statistics TO rtmc_user;

-- Create a table for tracking MainServer instances and their health
CREATE TABLE IF NOT EXISTS server_instances (
    instance_id SERIAL PRIMARY KEY,
    server_name VARCHAR(100) NOT NULL,
    server_ip VARCHAR(45) NOT NULL,
    server_port INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, inactive, maintenance
    connected_clients INTEGER DEFAULT 0,
    last_heartbeat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_server_instance UNIQUE (server_name, server_ip, server_port)
);

CREATE INDEX IF NOT EXISTS idx_server_instances_status ON server_instances(status);
CREATE INDEX IF NOT EXISTS idx_server_instances_heartbeat ON server_instances(last_heartbeat DESC);

COMMENT ON TABLE server_instances IS 'Tracks MainServer instances for load balancing and monitoring';
COMMENT ON COLUMN server_instances.status IS 'Server status: active, inactive, or maintenance';
COMMENT ON COLUMN server_instances.connected_clients IS 'Number of currently connected clients';
COMMENT ON COLUMN server_instances.last_heartbeat IS 'Last time server reported being alive';

GRANT ALL PRIVILEGES ON TABLE server_instances TO rtmc_user;
GRANT USAGE, SELECT ON SEQUENCE server_instances_instance_id_seq TO rtmc_user;
