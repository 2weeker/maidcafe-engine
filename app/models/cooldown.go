package models

import (
	"context"
	"fmt"
	"time"
)

type CooldownError struct {
	Wait time.Duration
	What string
}

func (e CooldownError) Error() string {
	secs := int(e.Wait.Seconds()) + 1
	if e.Wait >= time.Hour {
		return fmt.Sprintf("slow down: you can %s again in about %d hours", e.What, int(e.Wait.Hours())+1)
	}
	if e.Wait >= time.Minute {
		return fmt.Sprintf("slow down: you can %s again in %d minutes", e.What, int(e.Wait.Minutes())+1)
	}
	return fmt.Sprintf("slow down: you can %s again in %d seconds", e.What, secs)
}

func (s *Store) checkInterval(ctx context.Context, sql string, userID int64, every time.Duration, what string) error {
	if every <= 0 {
		return nil
	}
	var last *time.Time
	if err := s.DB.QueryRow(ctx, sql, userID).Scan(&last); err != nil {
		return err
	}
	if last == nil {
		return nil
	}
	if wait := every - time.Since(*last); wait > 0 {
		return CooldownError{Wait: wait, What: what}
	}
	return nil
}

func (s *Store) PostCooldown(ctx context.Context, userID int64, every time.Duration) error {
	return s.checkInterval(ctx, `SELECT max(created_at) FROM posts WHERE user_id = $1`, userID, every, "post")
}

func (s *Store) NoteCooldown(ctx context.Context, userID int64, every time.Duration) error {
	return s.checkInterval(ctx, `SELECT max(created_at) FROM notes WHERE user_id = $1`, userID, every, "post a note")
}

func (s *Store) ThreadQuota(ctx context.Context, userID int64, perDay int) error {
	if perDay <= 0 {
		return nil
	}
	var n int
	var oldest *time.Time
	if err := s.DB.QueryRow(ctx, `SELECT count(*), min(created_at) FROM threads
		WHERE user_id = $1 AND created_at > now() - interval '24 hours'`, userID).Scan(&n, &oldest); err != nil {
		return err
	}
	if n < perDay || oldest == nil {
		return nil
	}
	return CooldownError{Wait: 24*time.Hour - time.Since(*oldest), What: "start a thread"}
}
