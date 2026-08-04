
BEGIN;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DROP VIEW IF EXISTS recent_messages;
DROP VIEW IF EXISTS message_statistics;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                   WHERE table_schema = 'public' AND table_name = 'messages') THEN
        RAISE NOTICE 'messages table does not exist - run database/schema.sql instead';
        RETURN;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'messages' AND column_name = 'message_id' AND data_type = 'bigint') THEN
        ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_id_sub_id_idx;
        ALTER TABLE messages RENAME COLUMN message_id TO id_seq;
        ALTER TABLE messages RENAME COLUMN id TO user_id;
        ALTER TABLE messages RENAME COLUMN id_seq TO id;
        ALTER TABLE messages RENAME COLUMN message_content TO content;
        ALTER TABLE messages ADD COLUMN message_id UUID NOT NULL DEFAULT gen_random_uuid();
    END IF;

    ALTER TABLE messages DROP COLUMN IF EXISTS metadata;
    ALTER TABLE messages DROP COLUMN IF EXISTS priority;

    ALTER TABLE messages ADD COLUMN IF NOT EXISTS message_id UUID NOT NULL DEFAULT gen_random_uuid();
    ALTER TABLE messages ADD COLUMN IF NOT EXISTS user_id VARCHAR(255) NOT NULL DEFAULT '';
    ALTER TABLE messages ADD COLUMN IF NOT EXISTS sub_id VARCHAR(255) NOT NULL DEFAULT '';
    ALTER TABLE messages ADD COLUMN IF NOT EXISTS command VARCHAR(255) NOT NULL DEFAULT '';
    ALTER TABLE messages ADD COLUMN IF NOT EXISTS publisher_info TEXT NOT NULL DEFAULT '{}';
    ALTER TABLE messages ADD COLUMN IF NOT EXISTS server_name VARCHAR(100) NOT NULL DEFAULT 'MainServer';
    ALTER TABLE messages ADD COLUMN IF NOT EXISTS content TEXT NOT NULL DEFAULT '';
    ALTER TABLE messages ADD COLUMN IF NOT EXISTS is_encrypted BOOLEAN NOT NULL DEFAULT FALSE;
    ALTER TABLE messages ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending';
    ALTER TABLE messages ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP WITH TIME ZONE;

    ALTER TABLE messages ALTER COLUMN user_id TYPE VARCHAR(255);
    ALTER TABLE messages ALTER COLUMN sub_id TYPE VARCHAR(255);
    ALTER TABLE messages ALTER COLUMN command TYPE VARCHAR(255);

    UPDATE messages SET sub_id = '' WHERE sub_id IS NULL;
    ALTER TABLE messages ALTER COLUMN sub_id SET NOT NULL;
    ALTER TABLE messages ALTER COLUMN sub_id SET DEFAULT '';
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS messages_message_id_key ON messages(message_id);

DROP TABLE IF EXISTS message_history;
DROP TABLE IF EXISTS server_instances;

DELETE FROM encryption_keys WHERE key_name = 'default_key' AND encryption_key LIKE 'REPLACE_WITH%';

COMMIT;
