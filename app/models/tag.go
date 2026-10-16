package models

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrTagTaken   = errors.New("that tag already exists")
	ErrUnknownTag = errors.New("unknown tag")
	ErrTooManyTag = errors.New("too many tags")

	tagRe = regexp.MustCompile(`^[a-z0-9_-]+$`)
)

type Tag struct {
	ID          int64
	Name        string
	CreatedBy   string
	CreatedAt   time.Time
	NoteCount   int
	ThreadCount int
}

func (t Tag) Uses() int { return t.NoteCount + t.ThreadCount }

type TagSort string

const (
	TagSortPopular TagSort = "popular"
	TagSortNewest  TagSort = "newest"
	TagSortName    TagSort = "name"
)

func ParseTagSort(s string) TagSort {
	switch TagSort(s) {
	case TagSortNewest, TagSortName:
		return TagSort(s)
	}
	return TagSortPopular
}

func NormalizeTag(name string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(name), "#")))
}

type BadTagError struct{ MaxLen int }

func (e BadTagError) Error() string {
	return fmt.Sprintf("tags are 1-%d characters: a-z, 0-9, - and _", e.MaxLen)
}

func ValidateTag(name string, maxLen int) error {
	n := utf8.RuneCountInString(name)
	if n < 1 || n > maxLen || !tagRe.MatchString(name) {
		return BadTagError{MaxLen: maxLen}
	}
	return nil
}

const tagCols = `t.id, t.name, coalesce(u.username, ''), t.created_at,
	(SELECT count(*) FROM note_tags nt WHERE nt.tag_id = t.id),
	(SELECT count(*) FROM thread_tags tt WHERE tt.tag_id = t.id)
	FROM tags t LEFT JOIN users u ON u.id = t.created_by`

func (s *Store) queryTags(ctx context.Context, sql string, args ...any) ([]Tag, error) {
	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Tag, error) {
		var t Tag
		err := r.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt, &t.NoteCount, &t.ThreadCount)
		return t, err
	})
}

func (s *Store) CreateTag(ctx context.Context, userID int64, name string, maxLen int) (*Tag, error) {
	name = NormalizeTag(name)
	if err := ValidateTag(name, maxLen); err != nil {
		return nil, err
	}
	var id int64
	err := s.DB.QueryRow(ctx, `INSERT INTO tags (name, created_by) VALUES ($1, $2) RETURNING id`, name, userID).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrTagTaken
		}
		return nil, err
	}
	return s.TagByID(ctx, id)
}

func (s *Store) TagByID(ctx context.Context, id int64) (*Tag, error) {
	tags, err := s.queryTags(ctx, `SELECT `+tagCols+` WHERE t.id = $1`, id)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, ErrNotFound
	}
	return &tags[0], nil
}

func (s *Store) TagByName(ctx context.Context, name string) (*Tag, error) {
	tags, err := s.queryTags(ctx, `SELECT `+tagCols+` WHERE t.name = $1`, NormalizeTag(name))
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, ErrNotFound
	}
	return &tags[0], nil
}

func (s *Store) Tags(ctx context.Context, sort TagSort, limit, offset int) ([]Tag, error) {
	order := `(SELECT count(*) FROM note_tags nt WHERE nt.tag_id = t.id) + (SELECT count(*) FROM thread_tags tt WHERE tt.tag_id = t.id) DESC, t.name`
	switch sort {
	case TagSortNewest:
		order = `t.created_at DESC, t.id DESC`
	case TagSortName:
		order = `t.name`
	}
	return s.queryTags(ctx, `SELECT `+tagCols+` ORDER BY `+order+` LIMIT $1 OFFSET $2`, limit, offset)
}

func (s *Store) TagNames(ctx context.Context) ([]string, error) {
	rows, err := s.DB.Query(ctx, `SELECT name FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

func (s *Store) ResolveTags(ctx context.Context, names []string, max int) ([]int64, error) {
	var clean []string
	for _, n := range names {
		n = NormalizeTag(n)
		if n != "" && !slices.Contains(clean, n) {
			clean = append(clean, n)
		}
	}
	if len(clean) == 0 {
		return nil, nil
	}
	if len(clean) > max {
		return nil, fmt.Errorf("%w: at most %d per post", ErrTooManyTag, max)
	}
	rows, err := s.DB.Query(ctx, `SELECT id, name FROM tags WHERE name = ANY($1)`, clean)
	if err != nil {
		return nil, err
	}
	found := map[string]int64{}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			rows.Close()
			return nil, err
		}
		found[name] = id
	}
	rows.Close()
	ids := make([]int64, 0, len(clean))
	for _, n := range clean {
		id, ok := found[n]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTag, n)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func attachTags(ctx context.Context, tx pgx.Tx, table, column string, ownerID int64, tagIDs []int64) error {
	for _, id := range tagIDs {
		if _, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s (%s, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, table, column), ownerID, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) NotesByTag(ctx context.Context, viewer, tagID int64, sort Sort, limit, offset int) ([]FeedItem, error) {
	order := `n.created_at DESC, n.id DESC`
	if sort == SortTop {
		order = topOrder
	}
	return s.queryFeed(ctx, plainSQL(`EXISTS (SELECT 1 FROM note_tags nt WHERE nt.note_id = n.id AND nt.tag_id = $2)`, order)+
		` LIMIT $3 OFFSET $4`, viewer, tagID, limit, offset)
}

func (s *Store) ThreadsByTag(ctx context.Context, tagID int64, limit, offset int) ([]Thread, error) {
	return s.queryThreads(ctx, `SELECT `+threadCols+`
		WHERE EXISTS (SELECT 1 FROM thread_tags tt WHERE tt.thread_id = t.id AND tt.tag_id = $1)
		ORDER BY t.bumped_at DESC, t.id DESC LIMIT $2 OFFSET $3`, tagID, limit, offset)
}

func (s *Store) PopularTags(ctx context.Context, limit int) ([]Tag, error) {
	return s.Tags(ctx, TagSortPopular, limit, 0)
}
