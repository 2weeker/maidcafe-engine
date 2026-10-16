CREATE TABLE follows (
    follower_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    followee_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (follower_id, followee_id),
    CHECK (follower_id <> followee_id)
);
CREATE INDEX follows_followee_idx ON follows (followee_id, created_at DESC);
CREATE INDEX follows_follower_idx ON follows (follower_id, created_at DESC);

CREATE TABLE tags (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE CHECK (name ~ '^[a-z0-9_-]{1,32}$'),
    created_by BIGINT REFERENCES users (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE note_tags (
    note_id BIGINT NOT NULL REFERENCES notes (id) ON DELETE CASCADE,
    tag_id  BIGINT NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (note_id, tag_id)
);
CREATE INDEX note_tags_tag_idx ON note_tags (tag_id, note_id DESC);

CREATE TABLE thread_tags (
    thread_id BIGINT NOT NULL REFERENCES threads (id) ON DELETE CASCADE,
    tag_id    BIGINT NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (thread_id, tag_id)
);
CREATE INDEX thread_tags_tag_idx ON thread_tags (tag_id, thread_id DESC);

CREATE INDEX notes_body_fts_idx ON notes USING GIN (to_tsvector('simple', body));
CREATE INDEX posts_body_fts_idx ON posts USING GIN (to_tsvector('simple', body));
CREATE INDEX threads_title_fts_idx ON threads USING GIN (to_tsvector('simple', title));
CREATE INDEX users_username_pattern_idx ON users (username text_pattern_ops);
