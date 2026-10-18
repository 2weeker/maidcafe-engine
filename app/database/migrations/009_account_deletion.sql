ALTER TABLE users ADD COLUMN delete_at TIMESTAMPTZ;
CREATE INDEX users_delete_at_idx ON users (delete_at) WHERE delete_at IS NOT NULL;

ALTER TABLE threads
    ALTER COLUMN user_id DROP NOT NULL,
    DROP CONSTRAINT threads_user_id_fkey,
    ADD CONSTRAINT threads_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL;

ALTER TABLE posts
    ALTER COLUMN user_id DROP NOT NULL,
    DROP CONSTRAINT posts_user_id_fkey,
    ADD CONSTRAINT posts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL;
