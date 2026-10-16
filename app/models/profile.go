package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MaxBio      = 500
	MaxArtLines = 24
	MaxArtCols  = 80
)

var (
	ErrBadBio = fmt.Errorf("bio must be at most %d characters", MaxBio)
	ErrBadArt = errors.New("bad art")
)

func CleanBio(s string) (string, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
	if utf8.RuneCountInString(s) > MaxBio {
		return "", ErrBadBio
	}
	return s, nil
}

func CleanArt(s string) (string, error) {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\t", "    ")
	s = strings.Map(func(r rune) rune {
		if r != '\n' && unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)

	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRightFunc(l, unicode.IsSpace)
	}
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) > MaxArtLines {
		return "", fmt.Errorf("%w: art is %d lines tall (max %d)", ErrBadArt, len(lines), MaxArtLines)
	}
	for i, l := range lines {
		if n := utf8.RuneCountInString(l); n > MaxArtCols {
			return "", fmt.Errorf("%w: art line %d is %d columns wide (max %d)", ErrBadArt, i+1, n, MaxArtCols)
		}
	}
	return strings.Join(lines, "\n"), nil
}

func (s *Store) UpdateProfile(ctx context.Context, userID int64, bio, art string) error {
	bio, err := CleanBio(bio)
	if err != nil {
		return err
	}
	if art, err = CleanArt(art); err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx, `UPDATE users SET bio = $1, art = $2 WHERE id = $3`, bio, art, userID)
	return err
}
