package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrSelfFollow = errors.New("you can't follow yourself")

type UserCard struct {
	ID        int64
	Username  string
	Bio       string
	CreatedAt time.Time
	Followers int
	Following int
	Notes     int
	Followed  bool
}

type UserSort string

const (
	UserSortNewest   UserSort = "newest"
	UserSortFollowed UserSort = "followed"
	UserSortName     UserSort = "name"
)

func ParseUserSort(s string) UserSort {
	switch UserSort(s) {
	case UserSortFollowed, UserSortName:
		return UserSort(s)
	}
	return UserSortNewest
}

const userCardCols = `u.id, u.username, u.bio, u.created_at,
	(SELECT count(*) FROM follows f WHERE f.followee_id = u.id),
	(SELECT count(*) FROM follows f WHERE f.follower_id = u.id),
	(SELECT count(*) FROM notes n WHERE n.user_id = u.id AND n.reply_to_id IS NULL),
	EXISTS (SELECT 1 FROM follows f WHERE f.followee_id = u.id AND f.follower_id = $1)`

func (s *Store) queryUserCards(ctx context.Context, sql string, args ...any) ([]UserCard, error) {
	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (UserCard, error) {
		var c UserCard
		err := r.Scan(&c.ID, &c.Username, &c.Bio, &c.CreatedAt, &c.Followers, &c.Following, &c.Notes, &c.Followed)
		return c, err
	})
}

func (s *Store) UserCard(ctx context.Context, viewer, userID int64) (*UserCard, error) {
	cards, err := s.queryUserCards(ctx, `SELECT `+userCardCols+` FROM users u WHERE u.id = $2`, viewer, userID)
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, ErrNotFound
	}
	return &cards[0], nil
}

func (s *Store) Users(ctx context.Context, viewer int64, sort UserSort, limit, offset int) ([]UserCard, error) {
	switch sort {
	case UserSortFollowed:
		return s.Following(ctx, viewer, viewer, limit, offset)
	case UserSortName:
		return s.queryUserCards(ctx, `SELECT `+userCardCols+` FROM users u ORDER BY u.username LIMIT $2 OFFSET $3`,
			viewer, limit, offset)
	}
	return s.queryUserCards(ctx, `SELECT `+userCardCols+` FROM users u ORDER BY u.created_at DESC, u.id DESC LIMIT $2 OFFSET $3`,
		viewer, limit, offset)
}

func (s *Store) Followers(ctx context.Context, viewer, userID int64, limit, offset int) ([]UserCard, error) {
	return s.queryUserCards(ctx, `SELECT `+userCardCols+` FROM follows fl JOIN users u ON u.id = fl.follower_id
		WHERE fl.followee_id = $2 ORDER BY fl.created_at DESC LIMIT $3 OFFSET $4`, viewer, userID, limit, offset)
}

func (s *Store) Following(ctx context.Context, viewer, userID int64, limit, offset int) ([]UserCard, error) {
	return s.queryUserCards(ctx, `SELECT `+userCardCols+` FROM follows fl JOIN users u ON u.id = fl.followee_id
		WHERE fl.follower_id = $2 ORDER BY fl.created_at DESC LIMIT $3 OFFSET $4`, viewer, userID, limit, offset)
}

func (s *Store) SetFollow(ctx context.Context, followerID, followeeID int64, on bool) error {
	if followerID == followeeID {
		return ErrSelfFollow
	}
	var err error
	if on {
		_, err = s.DB.Exec(ctx, `INSERT INTO follows (follower_id, followee_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			followerID, followeeID)
	} else {
		_, err = s.DB.Exec(ctx, `DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`, followerID, followeeID)
	}
	return fkNotFound(err)
}

func (s *Store) NewestUsers(ctx context.Context, limit int) ([]UserCard, error) {
	return s.queryUserCards(ctx, `SELECT `+userCardCols+` FROM users u ORDER BY u.created_at DESC, u.id DESC LIMIT $2`, 0, limit)
}
