package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"
)

const SessionTTL = 30 * 24 * time.Hour

func (s *Store) CreateSession(ctx context.Context, userID int64) (string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	expires := time.Now().Add(SessionTTL)
	_, err := s.DB.Exec(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		hashToken(token), userID, expires,
	)
	return token, expires, err
}

func (s *Store) UserBySession(ctx context.Context, token string) (*User, error) {
	u := &User{}
	err := s.DB.QueryRow(ctx,
		`SELECT u.id, u.username, u.password_hash, u.bio, u.art, u.created_at, u.delete_at, u.role, u.banned_at, u.ban_reason
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = $1 AND s.expires_at > now()`,
		hashToken(token),
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Bio, &u.Art, &u.CreatedAt, &u.DeleteAt, &u.Role, &u.BannedAt, &u.BanReason)
	if err != nil {
		return nil, notFound(err)
	}
	return u, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, hashToken(token))
	return err
}

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
