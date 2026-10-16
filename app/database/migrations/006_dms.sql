CREATE TABLE conversations (
    id              BIGSERIAL PRIMARY KEY,
    user_a          BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    user_b          BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    last_read_a     BIGINT NOT NULL DEFAULT 0,
    last_read_b     BIGINT NOT NULL DEFAULT 0,
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_a, user_b),
    CHECK (user_a < user_b)
);
CREATE INDEX conversations_a_idx ON conversations (user_a, last_message_at DESC);
CREATE INDEX conversations_b_idx ON conversations (user_b, last_message_at DESC);

CREATE TABLE messages (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,
    sender_id       BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    body            TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 2000),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX messages_conversation_idx ON messages (conversation_id, id DESC);
