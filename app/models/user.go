package models

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrBadUsername   = errors.New("username must be 3-20 characters: a-z, 0-9, _")
	ErrBadPassword   = errors.New("password must be 8-72 characters")
	ErrUsernameTaken = errors.New("username is taken")
	ErrBadLogin      = errors.New("wrong username or password")

	usernameRe = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Bio          string
	Art          string
	CreatedAt    time.Time
}

func NormalizeUsername(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func (s *Store) CreateUser(ctx context.Context, username, password string) (*User, error) {
	username = NormalizeUsername(username)
	if !usernameRe.MatchString(username) {
		return nil, ErrBadUsername
	}
	if len(password) < 8 || len(password) > 72 {
		return nil, ErrBadPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{Username: username, PasswordHash: string(hash)}
	err = s.DB.QueryRow(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, created_at`,
		u.Username, u.PasswordHash,
	).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}
	return u, nil
}

func (s *Store) Authenticate(ctx context.Context, username, password string) (*User, error) {
	u, err := s.UserByUsername(ctx, username)
	if errors.Is(err, ErrNotFound) {

		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, ErrBadLogin
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, ErrBadLogin
	}
	return u, nil
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

func (s *Store) UserByUsername(ctx context.Context, username string) (*User, error) {
	u := &User{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, username, password_hash, bio, art, created_at FROM users WHERE username = $1`,
		NormalizeUsername(username),
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Bio, &u.Art, &u.CreatedAt)
	if err != nil {
		return nil, notFound(err)
	}
	return u, nil
}
