package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const DeletedName = "deleted"

var ErrWrongPassword = errors.New("wrong password")

func (s *Store) RequestDeletion(ctx context.Context, u *User, password string, grace time.Duration) (time.Time, error) {
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return time.Time{}, ErrWrongPassword
	}
	at := time.Now().Add(grace)
	_, err := s.DB.Exec(ctx, `UPDATE users SET delete_at = $2 WHERE id = $1`, u.ID, at)
	return at, err
}

func (s *Store) CancelDeletion(ctx context.Context, userID int64) error {
	_, err := s.DB.Exec(ctx, `UPDATE users SET delete_at = NULL WHERE id = $1`, userID)
	return err
}

func (s *Store) PurgeDeletedAccounts(ctx context.Context) ([]string, error) {
	var names []string
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `DELETE FROM users WHERE delete_at IS NOT NULL AND delete_at <= now() RETURNING username`)
		if err != nil {
			return err
		}
		names, err = pgx.CollectRows(rows, pgx.RowTo[string])
		if err != nil || len(names) == 0 {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE notes n SET
			like_count   = (SELECT count(*) FROM note_likes   l WHERE l.note_id = n.id),
			repost_count = (SELECT count(*) FROM note_reposts r WHERE r.note_id = n.id),
			reply_count  = (SELECT count(*) FROM notes c WHERE c.reply_to_id = n.id)`)
		return err
	})
	return names, err
}
