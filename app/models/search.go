package models

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
)

type SearchPost struct {
	Post
	ThreadTitle string
	BoardSlug   string
}

func CleanQuery(q string) string {
	q = strings.Join(strings.Fields(q), " ")
	if len(q) > 200 {
		q = q[:200]
	}
	return q
}

const tsq = `websearch_to_tsquery('simple', $2)`

func (s *Store) SearchNotes(ctx context.Context, viewer int64, q string, limit, offset int) ([]FeedItem, error) {
	return s.queryFeed(ctx, plainSQL(`to_tsvector('simple', n.body) @@ `+tsq,
		`ts_rank(to_tsvector('simple', n.body), `+tsq+`) DESC, n.created_at DESC, n.id DESC`)+
		` LIMIT $3 OFFSET $4`, viewer, q, limit, offset)
}

func (s *Store) SearchPosts(ctx context.Context, q string, limit, offset int) ([]SearchPost, error) {
	rows, err := s.DB.Query(ctx, `SELECT p.id, p.thread_id, p.no, u.username, p.body, p.sage, p.created_at, t.title, b.slug
		FROM posts p JOIN users u ON u.id = p.user_id JOIN threads t ON t.id = p.thread_id JOIN boards b ON b.id = t.board_id
		WHERE p.user_id IS NOT NULL AND p.deleted_at IS NULL AND to_tsvector('simple', p.body) @@ websearch_to_tsquery('simple', $1)
		ORDER BY ts_rank(to_tsvector('simple', p.body), websearch_to_tsquery('simple', $1)) DESC, p.created_at DESC, p.id DESC
		LIMIT $2 OFFSET $3`, q, limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (SearchPost, error) {
		var p SearchPost
		err := r.Scan(&p.ID, &p.ThreadID, &p.No, &p.Author, &p.Body, &p.Sage, &p.CreatedAt, &p.ThreadTitle, &p.BoardSlug)
		return p, err
	})
}

func (s *Store) SearchThreads(ctx context.Context, q string, limit, offset int) ([]Thread, error) {
	return s.queryThreads(ctx, `SELECT `+threadCols+`
		WHERE to_tsvector('simple', t.title) @@ websearch_to_tsquery('simple', $1) OR t.title ILIKE '%' || $1 || '%'
		ORDER BY t.bumped_at DESC, t.id DESC LIMIT $2 OFFSET $3`, q, limit, offset)
}

func (s *Store) SearchUsers(ctx context.Context, viewer int64, q string, limit, offset int) ([]UserCard, error) {
	return s.queryUserCards(ctx, `SELECT `+userCardCols+` FROM users u
		WHERE u.username LIKE '%' || $2 || '%' OR u.bio ILIKE '%' || $2 || '%'
		ORDER BY (u.username = $2) DESC, (u.username LIKE $2 || '%') DESC, u.username LIMIT $3 OFFSET $4`,
		viewer, NormalizeUsername(q), limit, offset)
}

func (s *Store) SearchTags(ctx context.Context, q string, limit, offset int) ([]Tag, error) {
	return s.queryTags(ctx, `SELECT `+tagCols+` WHERE t.name LIKE '%' || $1 || '%'
		ORDER BY (t.name = $1) DESC, (t.name LIKE $1 || '%') DESC, t.name LIMIT $2 OFFSET $3`, NormalizeTag(q), limit, offset)
}
