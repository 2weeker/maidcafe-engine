package models

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	RoleUser  = "user"
	RoleMod   = "mod"
	RoleAdmin = "admin"
)

var (
	ErrBadRole       = errors.New("role must be user, mod or admin")
	ErrBadBoard      = errors.New("board slug is 1-16 characters: a-z 0-9; name is required")
	ErrBoardTaken    = errors.New("a board with that slug already exists")
	ErrBoardNotEmpty = errors.New("that board still has threads")
	ErrThreadLocked  = errors.New("this thread is locked")
	ErrNotYours      = errors.New("that isn't yours to change")
	ErrGone          = errors.New("that post was deleted")

	boardSlugRe = regexp.MustCompile(`^[a-z0-9]{1,16}$`)
)

type ModAction struct {
	ID        int64
	Actor     string
	Action    string
	Target    string
	Detail    string
	CreatedAt time.Time
}

type AdminStats struct {
	Users, Banned, Mods, Notes, Threads, Posts, Tags, PendingDeletion int
}

type AdminUser struct {
	ID        int64
	Username  string
	Role      string
	BannedAt  *time.Time
	BanReason string
	DeleteAt  *time.Time
	CreatedAt time.Time
	Notes     int
	Posts     int
}

func (s *Store) Log(ctx context.Context, actorID int64, action, target, detail string) error {
	_, err := s.DB.Exec(ctx, `INSERT INTO mod_actions (actor_id, action, target, detail) VALUES ($1, $2, $3, $4)`,
		actorID, action, target, detail)
	return err
}

func (s *Store) ModLog(ctx context.Context, limit, offset int) ([]ModAction, error) {
	rows, err := s.DB.Query(ctx, `SELECT a.id, coalesce(u.username, 'system'), a.action, a.target, a.detail, a.created_at
		FROM mod_actions a LEFT JOIN users u ON u.id = a.actor_id
		ORDER BY a.id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (ModAction, error) {
		var a ModAction
		err := r.Scan(&a.ID, &a.Actor, &a.Action, &a.Target, &a.Detail, &a.CreatedAt)
		return a, err
	})
}

func (s *Store) Stats(ctx context.Context) (AdminStats, error) {
	var st AdminStats
	err := s.DB.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM users), (SELECT count(*) FROM users WHERE banned_at IS NOT NULL),
		(SELECT count(*) FROM users WHERE role <> 'user'), (SELECT count(*) FROM notes),
		(SELECT count(*) FROM threads), (SELECT count(*) FROM posts), (SELECT count(*) FROM tags),
		(SELECT count(*) FROM users WHERE delete_at IS NOT NULL)`).Scan(
		&st.Users, &st.Banned, &st.Mods, &st.Notes, &st.Threads, &st.Posts, &st.Tags, &st.PendingDeletion)
	return st, err
}

func (s *Store) AdminUsers(ctx context.Context, q string, limit, offset int) ([]AdminUser, error) {
	rows, err := s.DB.Query(ctx, `SELECT u.id, u.username, u.role, u.banned_at, u.ban_reason, u.delete_at, u.created_at,
			(SELECT count(*) FROM notes n WHERE n.user_id = u.id), (SELECT count(*) FROM posts p WHERE p.user_id = u.id)
		FROM users u WHERE $1 = '' OR u.username LIKE '%' || $1 || '%'
		ORDER BY (u.role <> 'user') DESC, (u.banned_at IS NOT NULL) DESC, u.created_at DESC LIMIT $2 OFFSET $3`,
		NormalizeUsername(q), limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (AdminUser, error) {
		var u AdminUser
		err := r.Scan(&u.ID, &u.Username, &u.Role, &u.BannedAt, &u.BanReason, &u.DeleteAt, &u.CreatedAt, &u.Notes, &u.Posts)
		return u, err
	})
}

func (s *Store) SyncAdmins(ctx context.Context, usernames []string) error {
	names := make([]string, 0, len(usernames))
	for _, n := range usernames {
		names = append(names, NormalizeUsername(n))
	}
	return pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE users SET role = 'user' WHERE role = 'admin' AND NOT (username = ANY($1))`, names); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE users SET role = 'admin', banned_at = NULL WHERE username = ANY($1)`, names)
		return err
	})
}

func (s *Store) SetRole(ctx context.Context, userID int64, role string) error {
	if role != RoleUser && role != RoleMod {
		return ErrBadRole
	}
	_, err := s.DB.Exec(ctx, `UPDATE users SET role = $2 WHERE id = $1 AND role <> 'admin'`, userID, role)
	return err
}

func (s *Store) SetBan(ctx context.Context, userID int64, on bool, reason string) error {
	reason = strings.TrimSpace(reason)
	if len(reason) > 200 {
		reason = reason[:200]
	}
	var err error
	if on {
		_, err = s.DB.Exec(ctx, `UPDATE users SET banned_at = now(), ban_reason = $2 WHERE id = $1 AND role <> 'admin'`, userID, reason)
	} else {
		_, err = s.DB.Exec(ctx, `UPDATE users SET banned_at = NULL, ban_reason = '' WHERE id = $1`, userID)
	}
	return err
}

func (s *Store) SetThreadLocked(ctx context.Context, threadID int64, on bool) error {
	_, err := s.DB.Exec(ctx, `UPDATE threads SET locked_at = CASE WHEN $2 THEN now() ELSE NULL END WHERE id = $1`, threadID, on)
	return err
}

func (s *Store) SetThreadPinned(ctx context.Context, threadID int64, on bool) error {
	_, err := s.DB.Exec(ctx, `UPDATE threads SET pinned = $2 WHERE id = $1`, threadID, on)
	return err
}

func (s *Store) DeleteThread(ctx context.Context, threadID int64) error {
	tag, err := s.DB.Exec(ctx, `DELETE FROM threads WHERE id = $1`, threadID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (s *Store) PostRaw(ctx context.Context, threadID int64, no int) (Post, error) {
	var p Post
	var userID *int64
	err := s.DB.QueryRow(ctx, `SELECT p.id, p.thread_id, p.no, coalesce(u.username, 'deleted'), p.body, p.sage, p.created_at,
			p.edited_at, p.deleted_at, p.removed, p.user_id
		FROM posts p LEFT JOIN users u ON u.id = p.user_id WHERE p.thread_id = $1 AND p.no = $2`, threadID, no).Scan(
		&p.ID, &p.ThreadID, &p.No, &p.Author, &p.Body, &p.Sage, &p.CreatedAt, &p.EditedAt, &p.DeletedAt, &p.Removed, &userID)
	if err != nil {
		return p, notFound(err)
	}
	if userID != nil {
		p.AuthorID = *userID
	}
	return p, nil
}

func (s *Store) EditPost(ctx context.Context, postID int64, body string) error {
	body, ok := cleanText(body, 4000)
	if !ok {
		return ErrBadBody
	}
	tag, err := s.DB.Exec(ctx, `UPDATE posts SET body = $2, edited_at = now() WHERE id = $1 AND deleted_at IS NULL`, postID, body)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrGone
	}
	return err
}

func (s *Store) DeletePost(ctx context.Context, postID int64, byMod bool) error {
	_, err := s.DB.Exec(ctx, `UPDATE posts SET deleted_at = now(), removed = $2 WHERE id = $1 AND deleted_at IS NULL`, postID, byMod)
	return err
}

func (s *Store) RestorePost(ctx context.Context, postID int64) error {
	_, err := s.DB.Exec(ctx, `UPDATE posts SET deleted_at = NULL, removed = false WHERE id = $1`, postID)
	return err
}

func (s *Store) NoteOwner(ctx context.Context, noteID int64) (int64, string, error) {
	var owner int64
	var body string
	err := s.DB.QueryRow(ctx, `SELECT user_id, body FROM notes WHERE id = $1`, noteID).Scan(&owner, &body)
	return owner, body, notFound(err)
}

func (s *Store) EditNote(ctx context.Context, noteID int64, body string) error {
	body, ok := cleanText(body, 300)
	if !ok {
		return ErrBadNote
	}
	_, err := s.DB.Exec(ctx, `UPDATE notes SET body = $2, edited_at = now() WHERE id = $1`, noteID, body)
	return err
}

func (s *Store) DeleteNote(ctx context.Context, noteID int64) error {
	return pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		var parent *int64
		if err := tx.QueryRow(ctx, `DELETE FROM notes WHERE id = $1 RETURNING reply_to_id`, noteID).Scan(&parent); err != nil {
			return notFound(err)
		}
		if parent != nil {
			if _, err := tx.Exec(ctx, `UPDATE notes SET reply_count = (SELECT count(*) FROM notes c WHERE c.reply_to_id = $1) WHERE id = $1`, *parent); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) CreateBoard(ctx context.Context, slug, name, description string, position int) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	name = strings.TrimSpace(name)
	if !boardSlugRe.MatchString(slug) || name == "" || len(name) > 40 {
		return ErrBadBoard
	}
	_, err := s.DB.Exec(ctx, `INSERT INTO boards (slug, name, description, position) VALUES ($1, $2, $3, $4)`,
		slug, name, strings.TrimSpace(description), position)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrBoardTaken
	}
	return err
}

func (s *Store) UpdateBoard(ctx context.Context, slug, name, description string, position int) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 40 {
		return ErrBadBoard
	}
	tag, err := s.DB.Exec(ctx, `UPDATE boards SET name = $2, description = $3, position = $4 WHERE slug = $1`,
		slug, name, strings.TrimSpace(description), position)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (s *Store) DeleteBoard(ctx context.Context, slug string) error {
	var threads int
	if err := s.DB.QueryRow(ctx, `SELECT count(*) FROM threads t JOIN boards b ON b.id = t.board_id WHERE b.slug = $1`, slug).Scan(&threads); err != nil {
		return err
	}
	if threads > 0 {
		return ErrBoardNotEmpty
	}
	tag, err := s.DB.Exec(ctx, `DELETE FROM boards WHERE slug = $1`, slug)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}
