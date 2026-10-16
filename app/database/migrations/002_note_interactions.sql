ALTER TABLE notes
    ADD COLUMN reply_to_id  BIGINT REFERENCES notes (id) ON DELETE CASCADE,
    ADD COLUMN like_count   INT NOT NULL DEFAULT 0,
    ADD COLUMN repost_count INT NOT NULL DEFAULT 0,
    ADD COLUMN reply_count  INT NOT NULL DEFAULT 0;

CREATE INDEX notes_reply_to_idx ON notes (reply_to_id, id) WHERE reply_to_id IS NOT NULL;
CREATE INDEX notes_toplevel_recent_idx ON notes (created_at DESC, id DESC) WHERE reply_to_id IS NULL;
CREATE INDEX notes_toplevel_top_idx ON notes (like_count DESC, created_at DESC, id DESC) WHERE reply_to_id IS NULL;

CREATE TABLE note_likes (
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    note_id    BIGINT NOT NULL REFERENCES notes (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, note_id)
);
CREATE INDEX note_likes_note_idx ON note_likes (note_id);

CREATE TABLE note_reposts (
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    note_id    BIGINT NOT NULL REFERENCES notes (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, note_id)
);
CREATE INDEX note_reposts_note_idx ON note_reposts (note_id);
CREATE INDEX note_reposts_recent_idx ON note_reposts (created_at DESC);
CREATE INDEX note_reposts_user_idx ON note_reposts (user_id, created_at DESC);
