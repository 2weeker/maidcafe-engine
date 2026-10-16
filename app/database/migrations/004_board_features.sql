ALTER TABLE posts
    ADD COLUMN no   INT,
    ADD COLUMN sage BOOLEAN NOT NULL DEFAULT false;

UPDATE posts p SET no = r.rn
FROM (SELECT id, row_number() OVER (PARTITION BY thread_id ORDER BY id) AS rn FROM posts) r
WHERE r.id = p.id;

ALTER TABLE posts ALTER COLUMN no SET NOT NULL;
CREATE UNIQUE INDEX posts_thread_no_idx ON posts (thread_id, no);

ALTER TABLE threads ADD COLUMN archived_at TIMESTAMPTZ;
DROP INDEX threads_board_bumped_idx;
CREATE INDEX threads_board_live_idx ON threads (board_id, bumped_at DESC, id DESC) WHERE archived_at IS NULL;
CREATE INDEX threads_board_archived_idx ON threads (board_id, archived_at DESC) WHERE archived_at IS NOT NULL;

CREATE TABLE post_refs (
    thread_id BIGINT NOT NULL REFERENCES threads (id) ON DELETE CASCADE,
    to_no     INT NOT NULL,
    from_no   INT NOT NULL,
    PRIMARY KEY (thread_id, to_no, from_no),
    CHECK (to_no < from_no)
);

INSERT INTO post_refs (thread_id, to_no, from_no)
SELECT DISTINCT p.thread_id, m[1]::int, p.no
FROM posts p, regexp_matches(p.body, '(?<!>)>>(\d{1,6})(?!\d)', 'g') AS m
WHERE m[1]::int BETWEEN 1 AND p.no - 1;
