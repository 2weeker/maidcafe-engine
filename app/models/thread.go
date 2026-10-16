package models

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

var (
	ErrBadTitle       = errors.New("title must be 1-120 characters")
	ErrBadBody        = errors.New("post must be 1-4000 characters")
	ErrThreadFull     = errors.New("this thread has reached its post limit. start a new one")
	ErrThreadArchived = errors.New("this thread is archived and can't take replies")
)

type Thread struct {
	ID         int64
	BoardID    int64
	BoardSlug  string
	Title      string
	Author     string
	ReplyCount int
	CreatedAt  time.Time
	BumpedAt   time.Time
	ArchivedAt *time.Time
	Tags       []string
}

func (t Thread) Posts() int { return t.ReplyCount + 1 }

type Post struct {
	ID        int64
	ThreadID  int64
	No        int
	Author    string
	Body      string
	Sage      bool
	CreatedAt time.Time
}

type ThreadPreview struct {
	Thread
	Shown   []Post
	Omitted int
}

func cleanText(s string, max int) (string, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
	n := utf8.RuneCountInString(s)
	return s, n >= 1 && n <= max
}

const threadCols = `t.id, t.board_id, b.slug, t.title, u.username, t.reply_count, t.created_at, t.bumped_at, t.archived_at,
	coalesce((SELECT array_agg(g.name ORDER BY g.name) FROM thread_tags tt JOIN tags g ON g.id = tt.tag_id WHERE tt.thread_id = t.id), '{}'::text[])
	FROM threads t JOIN users u ON u.id = t.user_id JOIN boards b ON b.id = t.board_id`

func scanThread(r pgx.Row) (Thread, error) {
	var t Thread
	err := r.Scan(&t.ID, &t.BoardID, &t.BoardSlug, &t.Title, &t.Author, &t.ReplyCount, &t.CreatedAt, &t.BumpedAt, &t.ArchivedAt, &t.Tags)
	return t, err
}

func (s *Store) queryThreads(ctx context.Context, sql string, args ...any) ([]Thread, error) {
	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Thread, error) { return scanThread(r) })
}

func (s *Store) LiveThreads(ctx context.Context, boardID int64, limit, offset int) ([]Thread, error) {
	return s.queryThreads(ctx, `SELECT `+threadCols+`
		WHERE t.board_id = $1 AND t.archived_at IS NULL
		ORDER BY t.bumped_at DESC, t.id DESC LIMIT $2 OFFSET $3`, boardID, limit, offset)
}

func (s *Store) ArchivedThreads(ctx context.Context, boardID int64, limit, offset int) ([]Thread, error) {
	return s.queryThreads(ctx, `SELECT `+threadCols+`
		WHERE t.board_id = $1 AND t.archived_at IS NOT NULL
		ORDER BY t.archived_at DESC, t.id DESC LIMIT $2 OFFSET $3`, boardID, limit, offset)
}

func (s *Store) ThreadByID(ctx context.Context, id int64) (*Thread, error) {
	t, err := scanThread(s.DB.QueryRow(ctx, `SELECT `+threadCols+` WHERE t.id = $1`, id))
	if err != nil {
		return nil, notFound(err)
	}
	return &t, nil
}

func (s *Store) CreateThread(ctx context.Context, boardID, userID int64, title, body string, maxThreads int, tagIDs []int64) (int64, error) {
	title, ok := cleanText(title, 120)
	if !ok {
		return 0, ErrBadTitle
	}
	body, ok = cleanText(body, 4000)
	if !ok {
		return 0, ErrBadBody
	}

	var id int64
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`INSERT INTO threads (board_id, user_id, title) VALUES ($1, $2, $3) RETURNING id`,
			boardID, userID, title).Scan(&id); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO posts (thread_id, user_id, body, no) VALUES ($1, $2, $3, 1)`, id, userID, body); err != nil {
			return err
		}
		if err := attachTags(ctx, tx, "thread_tags", "thread_id", id, tagIDs); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE threads SET archived_at = now() WHERE id IN (
			SELECT id FROM threads WHERE board_id = $1 AND archived_at IS NULL
			ORDER BY bumped_at DESC, id DESC OFFSET $2)`, boardID, maxThreads)
		return err
	})
	return id, err
}

type ReplyOptions struct {
	Sage      bool
	BumpLimit int
	PostLimit int
	Refs      []int
}

func (s *Store) CreateReply(ctx context.Context, threadID, userID int64, body string, opt ReplyOptions) (int, error) {
	body, ok := cleanText(body, 4000)
	if !ok {
		return 0, ErrBadBody
	}
	var no int
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {

		err := tx.QueryRow(ctx, `UPDATE threads SET
				reply_count = reply_count + 1,
				bumped_at = CASE WHEN $2 OR reply_count + 2 > $3 THEN bumped_at ELSE now() END
			WHERE id = $1 AND archived_at IS NULL AND reply_count + 1 < $4
			RETURNING reply_count + 1`,
			threadID, opt.Sage, opt.BumpLimit, opt.PostLimit).Scan(&no)
		if errors.Is(err, pgx.ErrNoRows) {
			return s.whyNoReply(ctx, tx, threadID)
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO posts (thread_id, user_id, body, no, sage) VALUES ($1, $2, $3, $4, $5)`,
			threadID, userID, body, no, opt.Sage); err != nil {
			return err
		}
		for _, to := range opt.Refs {
			if to < 1 || to >= no {
				continue
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO post_refs (thread_id, to_no, from_no) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
				threadID, to, no); err != nil {
				return err
			}
		}
		return nil
	})
	return no, err
}

func (s *Store) whyNoReply(ctx context.Context, tx pgx.Tx, threadID int64) error {
	var archived bool
	err := tx.QueryRow(ctx, `SELECT archived_at IS NOT NULL FROM threads WHERE id = $1`, threadID).Scan(&archived)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNotFound
	case err != nil:
		return err
	case archived:
		return ErrThreadArchived
	}
	return ErrThreadFull
}

const postCols = `p.id, p.thread_id, p.no, u.username, p.body, p.sage, p.created_at
	FROM posts p JOIN users u ON u.id = p.user_id`

func (s *Store) queryPosts(ctx context.Context, sql string, args ...any) ([]Post, error) {
	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Post, error) {
		var p Post
		err := r.Scan(&p.ID, &p.ThreadID, &p.No, &p.Author, &p.Body, &p.Sage, &p.CreatedAt)
		return p, err
	})
}

func (s *Store) PostsByThread(ctx context.Context, threadID int64, limit, offset int) ([]Post, error) {
	return s.queryPosts(ctx, `SELECT `+postCols+`
		WHERE p.thread_id = $1 ORDER BY p.no LIMIT $2 OFFSET $3`, threadID, limit, offset)
}

func (s *Store) Previews(ctx context.Context, threads []Thread, latest int) ([]ThreadPreview, error) {
	ids := make([]int64, len(threads))
	for i, t := range threads {
		ids[i] = t.ID
	}
	posts, err := s.queryPosts(ctx, `SELECT `+postCols+`
		JOIN threads t ON t.id = p.thread_id
		WHERE p.thread_id = ANY($1) AND (p.no = 1 OR p.no > t.reply_count + 1 - $2)
		ORDER BY p.thread_id, p.no`, ids, latest)
	if err != nil {
		return nil, err
	}
	byThread := make(map[int64][]Post, len(threads))
	for _, p := range posts {
		byThread[p.ThreadID] = append(byThread[p.ThreadID], p)
	}
	out := make([]ThreadPreview, len(threads))
	for i, t := range threads {
		out[i] = ThreadPreview{Thread: t, Shown: byThread[t.ID], Omitted: max(0, t.ReplyCount-latest)}
	}
	return out, nil
}

func (s *Store) Backlinks(ctx context.Context, threadID int64, fromNo, toNo int) (map[int][]int, error) {
	rows, err := s.DB.Query(ctx, `SELECT to_no, from_no FROM post_refs
		WHERE thread_id = $1 AND to_no BETWEEN $2 AND $3 ORDER BY to_no, from_no`, threadID, fromNo, toNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int][]int{}
	for rows.Next() {
		var to, from int
		if err := rows.Scan(&to, &from); err != nil {
			return nil, err
		}
		out[to] = append(out[to], from)
	}
	return out, rows.Err()
}
