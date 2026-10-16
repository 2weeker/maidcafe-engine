package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func load(t *testing.T, yaml string) (Config, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", path)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("LISTEN_ADDR", "")
	t.Setenv("COOKIE_SECURE", "")
	return Load()
}

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	t.Setenv("CONFIG_PATH", filepath.Join(t.TempDir(), "nope.yaml"))
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Boards.Defaults != Defaults().Boards.Defaults {
		t.Fatalf("got %+v", cfg.Boards.Defaults)
	}
}

func TestLoadEmptyFileUsesDefaults(t *testing.T) {
	cfg, err := load(t, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Pagination != Defaults().Pagination {
		t.Fatalf("got %+v", cfg.Pagination)
	}
}

func TestPartialFileKeepsOtherDefaults(t *testing.T) {
	cfg, err := load(t, "boards:\n  defaults:\n    bump_limit: 50\n")
	if err != nil {
		t.Fatal(err)
	}
	got := cfg.Boards.Defaults
	if got.BumpLimit != 50 || got.PostLimit != 1000 || got.MaxThreads != 150 {
		t.Fatalf("got %+v", got)
	}
}

func TestOverridesMergeOntoDefaults(t *testing.T) {
	cfg, err := load(t, `
boards:
  defaults:
    bump_limit: 300
    post_limit: 1000
  overrides:
    random:
      bump_limit: 100
      preview_replies: 0
`)
	if err != nil {
		t.Fatal(err)
	}
	r := cfg.Boards.For("random")
	if r.BumpLimit != 100 || r.PostLimit != 1000 || r.PreviewReplies != 0 || r.MaxThreads != 150 {
		t.Fatalf("random: %+v", r)
	}
	if cfg.Boards.For("tech") != cfg.Boards.Defaults {
		t.Fatalf("tech should use defaults")
	}
}

func TestEnvOverridesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(path, []byte("server:\n  listen_addr: \":1\"\ndatabase:\n  url: file-url\n"), 0o600)
	t.Setenv("CONFIG_PATH", path)
	t.Setenv("DATABASE_URL", "env-url")
	t.Setenv("LISTEN_ADDR", "")
	t.Setenv("COOKIE_SECURE", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.URL != "env-url" || cfg.Server.ListenAddr != ":1" || !cfg.Server.CookieSecure {
		t.Fatalf("got %+v", cfg)
	}
}

func TestInvalidConfigs(t *testing.T) {
	tests := map[string]struct{ yaml, want string }{
		"unknown key":          {"boards:\n  defaults:\n    bump_limt: 5\n", "bump_limt"},
		"bump over post limit": {"boards:\n  defaults:\n    bump_limit: 2000\n", "can't exceed post_limit"},
		"bad override":         {"boards:\n  overrides:\n    tech:\n      max_threads: 0\n", "boards.overrides.tech.max_threads"},
		"zero page size":       {"pagination:\n  posts: 0\n", "pagination.posts"},
		"tag too long":         {"tags:\n  max_length: 40\n", "tags.max_length"},
		"negative tags":        {"tags:\n  max_per_post: -1\n", "tags.max_per_post"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := load(t, tt.yaml)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestTagDefaults(t *testing.T) {
	cfg, err := load(t, "tags:\n  max_length: 12\n")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Tags.MaxLength != 12 || cfg.Tags.MaxPerPost != 3 {
		t.Fatalf("got %+v", cfg.Tags)
	}
}
