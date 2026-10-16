package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrBadNote = errors.New("note must be 1-300 characters")

type Note struct {
	ID            int64
	Author        string
	Body          string
	CreatedAt     time.Time
	ReplyToID     *int64
	ReplyToAuthor string
	LikeCount     int
	RepostCount   int
	ReplyCount    int
	Liked         bool
	Reposted      bool
	Tags          []string
}

type FeedItem struct {
	Note
	RepostedBy string
}

type Sort string

const (
	SortRecent    Sort = "recent"
	SortTop       Sort = "top"
	SortFollowing Sort = "following"
)

func ParseSort(s string) Sort {
	switch Sort(s) {
	case SortTop, SortFollowing:
		return Sort(s)
	}
	return SortRecent
}

const noteCols = `n.id, u.username, n.body, n.created_at, n.reply_to_id, coalesce(pu.username, ''),
	n.like_count, n.repost_count, n.reply_count,
	EXISTS (SELECT 1 FROM note_likes   l WHERE l.note_id = n.id AND l.user_id = $1),
	EXISTS (SELECT 1 FROM note_reposts r WHERE r.note_id = n.id AND r.user_id = $1),
	coalesce((SELECT array_agg(t.name ORDER BY t.name) FROM note_tags nt JOIN tags t ON t.id = nt.tag_id WHERE nt.note_id = n.id), '{}'::text[])`

const noteJoins = `JOIN users u ON u.id = n.user_id
	LEFT JOIN notes p ON p.id = n.reply_to_id
	LEFT JOIN users pu ON pu.id = p.user_id`

func feedSQL(source string) string {
	return `SELECT coalesce(ru.username, ''), ` + noteCols + `
		FROM (` + source + `) f
		JOIN notes n ON n.id = f.note_id
		` + noteJoins + `
		LEFT JOIN users ru ON ru.id = f.reposter_id
		ORDER BY f.at DESC, f.note_id DESC, ru.username`
}

func plainSQL(where, order string) string {
	return `SELECT '', ` + noteCols + ` FROM notes n ` + noteJoins + ` WHERE ` + where + ` ORDER BY ` + order
}

const topOrder = `n.like_count DESC, n.created_at DESC, n.id DESC`

func (s *Store) Feed(ctx context.Context, viewer int64, sort Sort, limit, offset int) ([]FeedItem, error) {
	switch sort {
	case SortTop:
		return s.queryFeed(ctx, plainSQL(`n.reply_to_id IS NULL`, topOrder)+` LIMIT $2 OFFSET $3`,
			viewer, limit, offset)
	case SortFollowing:
		if viewer == 0 {
			return nil, nil
		}
		return s.queryFeed(ctx, feedSQL(`
			SELECT n.id AS note_id, NULL::bigint AS reposter_id, n.created_at AS at FROM notes n
			WHERE n.reply_to_id IS NULL AND (n.user_id = $1 OR n.user_id IN (SELECT followee_id FROM follows WHERE follower_id = $1))
			UNION ALL
			SELECT r.note_id, r.user_id, r.created_at FROM note_reposts r
			WHERE r.user_id = $1 OR r.user_id IN (SELECT followee_id FROM follows WHERE follower_id = $1)`)+` LIMIT $2 OFFSET $3`,
			viewer, limit, offset)
	}
	return s.queryFeed(ctx, feedSQL(`
		SELECT id AS note_id, NULL::bigint AS reposter_id, created_at AS at FROM notes WHERE reply_to_id IS NULL
		UNION ALL
		SELECT note_id, user_id, created_at FROM note_reposts`)+` LIMIT $2 OFFSET $3`,
		viewer, limit, offset)
}

func (s *Store) UserFeed(ctx context.Context, viewer, userID int64, sort Sort, limit, offset int) ([]FeedItem, error) {
	if sort == SortTop {
		return s.queryFeed(ctx, plainSQL(`n.user_id = $2`, topOrder)+` LIMIT $3 OFFSET $4`,
			viewer, userID, limit, offset)
	}
	return s.queryFeed(ctx, feedSQL(`
		SELECT id AS note_id, NULL::bigint AS reposter_id, created_at AS at FROM notes WHERE user_id = $2
		UNION ALL
		SELECT note_id, user_id, created_at FROM note_reposts WHERE user_id = $2`)+` LIMIT $3 OFFSET $4`,
		viewer, userID, limit, offset)
}

func (s *Store) Replies(ctx context.Context, viewer, noteID int64, limit, offset int) ([]FeedItem, error) {
	return s.queryFeed(ctx, plainSQL(`n.reply_to_id = $2`, `n.id`)+` LIMIT $3 OFFSET $4`,
		viewer, noteID, limit, offset)
}

func (s *Store) NoteByID(ctx context.Context, viewer, id int64) (*Note, error) {
	items, err := s.queryFeed(ctx, plainSQL(`n.id = $2`, `n.id`), viewer, id)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return &items[0].Note, nil
}

func (s *Store) queryFeed(ctx context.Context, sql string, args ...any) ([]FeedItem, error) {
	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (FeedItem, error) {
		var f FeedItem
		err := r.Scan(&f.RepostedBy, &f.ID, &f.Author, &f.Body, &f.CreatedAt, &f.ReplyToID, &f.ReplyToAuthor,
			&f.LikeCount, &f.RepostCount, &f.ReplyCount, &f.Liked, &f.Reposted, &f.Tags)
		return f, err
	})
}

func (s *Store) CreateNote(ctx context.Context, userID int64, body string, replyTo *int64, tagIDs []int64) (int64, error) {
	body, ok := cleanText(body, 300)
	if !ok {
		return 0, ErrBadNote
	}
	var id int64
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`INSERT INTO notes (user_id, body, reply_to_id) VALUES ($1, $2, $3) RETURNING id`,
			userID, body, replyTo).Scan(&id); err != nil {
			return err
		}
		if replyTo == nil {
			return attachTags(ctx, tx, "note_tags", "note_id", id, tagIDs)
		}
		_, err := tx.Exec(ctx, `UPDATE notes SET reply_count = reply_count + 1 WHERE id = $1`, *replyTo)
		return err
	})
	return id, fkNotFound(err)
}

func (s *Store) SetLike(ctx context.Context, userID, noteID int64, on bool) error {
	return s.setFlag(ctx, "note_likes", "like_count", userID, noteID, on)
}

func (s *Store) SetRepost(ctx context.Context, userID, noteID int64, on bool) error {
	return s.setFlag(ctx, "note_reposts", "repost_count", userID, noteID, on)
}

func (s *Store) setFlag(ctx context.Context, table, counter string, userID, noteID int64, on bool) error {
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		var tag pgconn.CommandTag
		var err error
		delta := 1
		if on {
			tag, err = tx.Exec(ctx, fmt.Sprintf(
				`INSERT INTO %s (user_id, note_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, table), userID, noteID)
		} else {
			delta = -1
			tag, err = tx.Exec(ctx, fmt.Sprintf(
				`DELETE FROM %s WHERE user_id = $1 AND note_id = $2`, table), userID, noteID)
		}
		if err != nil || tag.RowsAffected() == 0 {
			return err
		}
		_, err = tx.Exec(ctx, fmt.Sprintf(
			`UPDATE notes SET %[1]s = %[1]s + $1 WHERE id = $2`, counter), delta, noteID)
		return err
	})
	return fkNotFound(err)
}

func fkNotFound(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrNotFound
	}
	return err
}
