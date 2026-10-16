package models

import (
	"errors"
	"strings"
	"testing"
)

func TestCleanArt(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "empty", in: "", want: ""},
		{name: "only blank lines", in: "\n  \n\n", want: ""},
		{
			name: "keeps leading indentation of first line",
			in:   "   /\\_/\\\n  ( o.o )\n   > ^ <",
			want: "   /\\_/\\\n  ( o.o )\n   > ^ <",
		},
		{
			name: "trims blank lines around but not inside",
			in:   "\n\n ab\n\n cd \n\n",
			want: " ab\n\n cd",
		},
		{name: "crlf normalized", in: "a\r\nb", want: "a\nb"},
		{name: "tabs become four spaces", in: "a\tb", want: "a    b"},
		{name: "control chars and ansi escapes stripped", in: "\x1b[31mred\x1b[0m\x00!", want: "[31mred[0m!"},
		{name: "unicode block art counts runes", in: strings.Repeat("█", 80), want: strings.Repeat("█", 80)},
		{name: "too wide", in: "ok\n" + strings.Repeat("x", 81), wantErr: true},
		{name: "too many lines", in: strings.Repeat("x\n", 25), wantErr: true},
		{name: "24 lines is fine", in: strings.TrimSuffix(strings.Repeat("x\n", 24), "\n"), want: strings.TrimSuffix(strings.Repeat("x\n", 24), "\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CleanArt(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got %q", got)
				}
				if !errors.Is(err, ErrBadArt) {
					t.Fatalf("want ErrBadArt, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCleanArtErrorNamesLine(t *testing.T) {
	_, err := CleanArt("ok\nok\n" + strings.Repeat("x", 93))
	if err == nil || !strings.Contains(err.Error(), "line 3 is 93 columns") {
		t.Fatalf("got %v", err)
	}
}

func TestCleanBio(t *testing.T) {
	if got, err := CleanBio("  hi\r\nthere  "); err != nil || got != "hi\nthere" {
		t.Fatalf("got %q, %v", got, err)
	}
	if got, err := CleanBio("   "); err != nil || got != "" {
		t.Fatalf("empty bio should be allowed, got %q, %v", got, err)
	}
	if _, err := CleanBio(strings.Repeat("é", 501)); !errors.Is(err, ErrBadBio) {
		t.Fatalf("want ErrBadBio, got %v", err)
	}
}
