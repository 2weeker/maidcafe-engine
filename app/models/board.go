package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Board struct {
	ID           int64
	Slug         string
	Name         string
	Description  string
	Position     int
	ThreadCount  int
	Archived     int
	PostCount    int
	LastActivity *time.Time
}

func (s *Store) Boards(ctx context.Context) ([]Board, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT b.id, b.slug, b.name, b.description, b.position,
		        (SELECT count(*) FROM threads t WHERE t.board_id = b.id AND t.archived_at IS NULL),
		        (SELECT count(*) FROM threads t WHERE t.board_id = b.id AND t.archived_at IS NOT NULL),
		        (SELECT coalesce(sum(t.reply_count + 1), 0) FROM threads t WHERE t.board_id = b.id),
		        (SELECT max(t.bumped_at) FROM threads t WHERE t.board_id = b.id)
		 FROM boards b ORDER BY b.position, b.id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Board, error) {
		var b Board
		err := r.Scan(&b.ID, &b.Slug, &b.Name, &b.Description, &b.Position, &b.ThreadCount, &b.Archived, &b.PostCount, &b.LastActivity)
		return b, err
	})
}

func (s *Store) BoardBySlug(ctx context.Context, slug string) (*Board, error) {
	b := &Board{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, slug, name, description FROM boards WHERE slug = $1`, slug,
	).Scan(&b.ID, &b.Slug, &b.Name, &b.Description)
	if err != nil {
		return nil, notFound(err)
	}
	return b, nil
}
