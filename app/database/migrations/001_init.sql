CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE CHECK (username ~ '^[a-z0-9_]{3,20}$'),
    password_hash TEXT NOT NULL,
    bio           TEXT NOT NULL DEFAULT '' CHECK (char_length(bio) <= 500),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    token_hash BYTEA PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX sessions_user_id_idx ON sessions (user_id);

CREATE TABLE boards (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]{1,16}$'),
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    position    INT NOT NULL DEFAULT 0
);

CREATE TABLE threads (
    id          BIGSERIAL PRIMARY KEY,
    board_id    BIGINT NOT NULL REFERENCES boards (id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title       TEXT NOT NULL CHECK (char_length(title) BETWEEN 1 AND 120),
    reply_count INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    bumped_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX threads_board_bumped_idx ON threads (board_id, bumped_at DESC);

CREATE TABLE posts (
    id         BIGSERIAL PRIMARY KEY,
    thread_id  BIGINT NOT NULL REFERENCES threads (id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    body       TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 4000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX posts_thread_idx ON posts (thread_id, id);

CREATE TABLE notes (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    body       TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 300),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX notes_created_idx ON notes (created_at DESC);
CREATE INDEX notes_user_idx ON notes (user_id, created_at DESC);

INSERT INTO boards (slug, name, description, position) VALUES
    ('meta',    'Meta',    'About this place.',          0),
    ('general', 'General', 'Anything goes. Mostly.',     1),
    ('tech',    'Tech',    'Computers, code, hardware.', 2),
    ('random',  'Random',  'The void stares back.',      3);
