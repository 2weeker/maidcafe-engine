ALTER TABLE users
    ADD COLUMN role       TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'mod', 'admin')),
    ADD COLUMN banned_at  TIMESTAMPTZ,
    ADD COLUMN ban_reason TEXT NOT NULL DEFAULT '';

ALTER TABLE posts
    ADD COLUMN edited_at  TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN removed    BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE notes ADD COLUMN edited_at TIMESTAMPTZ;

ALTER TABLE threads
    ADD COLUMN locked_at TIMESTAMPTZ,
    ADD COLUMN pinned    BOOLEAN NOT NULL DEFAULT false;
DROP INDEX threads_board_live_idx;
CREATE INDEX threads_board_live_idx ON threads (board_id, pinned DESC, bumped_at DESC, id DESC) WHERE archived_at IS NULL;

CREATE TABLE mod_actions (
    id         BIGSERIAL PRIMARY KEY,
    actor_id   BIGINT REFERENCES users (id) ON DELETE SET NULL,
    action     TEXT NOT NULL,
    target     TEXT NOT NULL,
    detail     TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX mod_actions_created_idx ON mod_actions (created_at DESC);
