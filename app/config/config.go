package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Server     Server     `yaml:"server"`
	Database   Database   `yaml:"database"`
	Pagination Pagination `yaml:"pagination"`
	Boards     Boards     `yaml:"boards"`
	Tags       Tags       `yaml:"tags"`
	Sidebar    Sidebar    `yaml:"sidebar"`
	Chat       Chat       `yaml:"chat"`
}

type Chat struct {
	RetentionHours int `yaml:"retention_hours"`
}

type Sidebar struct {
	Boards int `yaml:"boards"`
	Users  int `yaml:"users"`
	Tags   int `yaml:"tags"`
}

type Tags struct {
	MaxLength  int `yaml:"max_length"`
	MaxPerPost int `yaml:"max_per_post"`
}

type Server struct {
	ListenAddr   string `yaml:"listen_addr"`
	CookieSecure bool   `yaml:"cookie_secure"`
}

type Database struct {
	URL string `yaml:"url"`
}

type Pagination struct {
	Notes    int `yaml:"notes"`
	Replies  int `yaml:"replies"`
	Threads  int `yaml:"threads"`
	Posts    int `yaml:"posts"`
	Archive  int `yaml:"archive"`
	Users    int `yaml:"users"`
	Tags     int `yaml:"tags"`
	Search   int `yaml:"search"`
	Messages int `yaml:"messages"`
}

type BoardRules struct {
	BumpLimit      int `yaml:"bump_limit"`
	PostLimit      int `yaml:"post_limit"`
	MaxThreads     int `yaml:"max_threads"`
	PreviewReplies int `yaml:"preview_replies"`
}

type RulesOverride struct {
	BumpLimit      *int `yaml:"bump_limit"`
	PostLimit      *int `yaml:"post_limit"`
	MaxThreads     *int `yaml:"max_threads"`
	PreviewReplies *int `yaml:"preview_replies"`
}

type Boards struct {
	Defaults  BoardRules               `yaml:"defaults"`
	Overrides map[string]RulesOverride `yaml:"overrides"`
}

func (b Boards) For(slug string) BoardRules {
	r := b.Defaults
	o, ok := b.Overrides[slug]
	if !ok {
		return r
	}
	set := func(dst *int, src *int) {
		if src != nil {
			*dst = *src
		}
	}
	set(&r.BumpLimit, o.BumpLimit)
	set(&r.PostLimit, o.PostLimit)
	set(&r.MaxThreads, o.MaxThreads)
	set(&r.PreviewReplies, o.PreviewReplies)
	return r
}

func Defaults() Config {
	return Config{
		Server:     Server{ListenAddr: ":8080"},
		Database:   Database{URL: "postgres://maidcafe:maidcafe@localhost:5432/maidcafe?sslmode=disable"},
		Pagination: Pagination{Notes: 30, Replies: 50, Threads: 10, Posts: 50, Archive: 50, Users: 30, Tags: 60, Search: 20, Messages: 50},
		Tags:       Tags{MaxLength: 8, MaxPerPost: 3},
		Sidebar:    Sidebar{Boards: 2, Users: 5, Tags: 12},
		Chat:       Chat{RetentionHours: 24},
		Boards: Boards{Defaults: BoardRules{
			BumpLimit:      300,
			PostLimit:      1000,
			MaxThreads:     150,
			PreviewReplies: 5,
		}},
	}
}

func Load() (Config, error) {
	path := env("CONFIG_PATH", "config.yaml")
	cfg := Defaults()

	f, err := os.Open(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		log.Printf("config: %s not found, using defaults", path)
	case err != nil:
		return cfg, err
	default:
		defer f.Close()
		if cfg, err = parse(f, cfg); err != nil {
			return cfg, fmt.Errorf("config %s: %w", path, err)
		}
		log.Printf("config: loaded %s", path)
	}

	cfg.Server.ListenAddr = env("LISTEN_ADDR", cfg.Server.ListenAddr)
	cfg.Database.URL = env("DATABASE_URL", cfg.Database.URL)
	if v, ok := os.LookupEnv("COOKIE_SECURE"); ok && v != "" {
		cfg.Server.CookieSecure = v == "true"
	}
	return cfg, cfg.validate()
}

func parse(r io.Reader, base Config) (Config, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&base); err != nil && !errors.Is(err, io.EOF) {
		return base, err
	}
	return base, nil
}

func (c Config) validate() error {
	p := c.Pagination
	for name, v := range map[string]int{
		"pagination.notes": p.Notes, "pagination.replies": p.Replies, "pagination.threads": p.Threads,
		"pagination.posts": p.Posts, "pagination.archive": p.Archive, "pagination.users": p.Users,
		"pagination.tags": p.Tags, "pagination.search": p.Search, "pagination.messages": p.Messages,
	} {
		if v < 1 {
			return fmt.Errorf("config: %s must be at least 1", name)
		}
	}
	for name, v := range map[string]int{"sidebar.boards": c.Sidebar.Boards, "sidebar.users": c.Sidebar.Users, "sidebar.tags": c.Sidebar.Tags} {
		if v < 0 {
			return fmt.Errorf("config: %s can't be negative", name)
		}
	}
	if c.Chat.RetentionHours < 0 {
		return fmt.Errorf("config: chat.retention_hours can't be negative (0 keeps messages forever)")
	}
	switch {
	case c.Tags.MaxLength < 1 || c.Tags.MaxLength > 32:
		return fmt.Errorf("config: tags.max_length must be between 1 and 32")
	case c.Tags.MaxPerPost < 0 || c.Tags.MaxPerPost > 20:
		return fmt.Errorf("config: tags.max_per_post must be between 0 and 20")
	}
	if err := c.Boards.Defaults.validate("boards.defaults"); err != nil {
		return err
	}
	for slug := range c.Boards.Overrides {
		if err := c.Boards.For(slug).validate("boards.overrides." + slug); err != nil {
			return err
		}
	}
	return nil
}

func (r BoardRules) validate(where string) error {
	switch {
	case r.BumpLimit < 1:
		return fmt.Errorf("config: %s.bump_limit must be at least 1", where)
	case r.PostLimit < 2:
		return fmt.Errorf("config: %s.post_limit must be at least 2", where)
	case r.BumpLimit > r.PostLimit:
		return fmt.Errorf("config: %s.bump_limit (%d) can't exceed post_limit (%d)", where, r.BumpLimit, r.PostLimit)
	case r.MaxThreads < 1:
		return fmt.Errorf("config: %s.max_threads must be at least 1", where)
	case r.PreviewReplies < 0:
		return fmt.Errorf("config: %s.preview_replies can't be negative", where)
	}
	return nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
