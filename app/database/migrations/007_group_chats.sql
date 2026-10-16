ALTER TABLE conversations
    ADD COLUMN kind TEXT NOT NULL DEFAULT 'direct' CHECK (kind IN ('direct', 'group')),
    ADD COLUMN name TEXT CHECK (name ~ '^[a-z0-9][a-z0-9_-]{0,23}$'),
    ADD COLUMN visibility TEXT NOT NULL DEFAULT 'private' CHECK (visibility IN ('public', 'private')),
    ADD COLUMN created_by BIGINT REFERENCES users (id) ON DELETE SET NULL,
    ALTER COLUMN user_a DROP NOT NULL,
    ALTER COLUMN user_b DROP NOT NULL,
    ADD CONSTRAINT conversations_shape CHECK (
        (kind = 'direct' AND user_a IS NOT NULL AND user_b IS NOT NULL AND name IS NULL) OR
        (kind = 'group' AND user_a IS NULL AND user_b IS NULL AND name IS NOT NULL));

CREATE UNIQUE INDEX conversations_group_name_idx ON conversations (name) WHERE kind = 'group';
CREATE INDEX conversations_public_idx ON conversations (last_message_at DESC) WHERE kind = 'group' AND visibility = 'public';

CREATE TABLE chat_members (
    conversation_id BIGINT NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    state           TEXT NOT NULL CHECK (state IN ('joined', 'left', 'invited')),
    last_read       BIGINT NOT NULL DEFAULT 0,
    invited_by      BIGINT REFERENCES users (id) ON DELETE SET NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (conversation_id, user_id)
);
CREATE INDEX chat_members_user_idx ON chat_members (user_id, state);

INSERT INTO chat_members (conversation_id, user_id, state, last_read)
SELECT id, user_a, 'joined', last_read_a FROM conversations
UNION ALL
SELECT id, user_b, 'joined', last_read_b FROM conversations;

DROP INDEX conversations_a_idx;
DROP INDEX conversations_b_idx;
ALTER TABLE conversations DROP COLUMN last_read_a, DROP COLUMN last_read_b;

ALTER TABLE messages
    ADD COLUMN kind TEXT NOT NULL DEFAULT 'msg' CHECK (kind IN ('msg', 'join', 'leave', 'invite', 'create')),
    ADD COLUMN target_id BIGINT REFERENCES users (id) ON DELETE SET NULL,
    DROP CONSTRAINT messages_body_check,
    ADD CONSTRAINT messages_body_check CHECK (kind <> 'msg' OR char_length(body) BETWEEN 1 AND 2000);
