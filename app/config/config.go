package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

const (
	ModeCafe      = "cafe"
	ModeTextboard = "textboard"
)

type Config struct {
	Mode         string       `yaml:"mode"`
	Server       Server       `yaml:"server"`
	Database     Database     `yaml:"database"`
	Pagination   Pagination   `yaml:"pagination"`
	Boards       Boards       `yaml:"boards"`
	Tags         Tags         `yaml:"tags"`
	Sidebar      Sidebar      `yaml:"sidebar"`
	Accounts     Accounts     `yaml:"accounts"`
	Admin        Admin        `yaml:"admin"`
	Cooldowns    Cooldowns    `yaml:"cooldowns"`
	Guests       Guests       `yaml:"guests"`
	Captcha      Captcha      `yaml:"captcha"`
	RateLimits   RateLimits   `yaml:"rate_limits"`
	Registration Registration `yaml:"registration"`
	About        About        `yaml:"about"`
}

type Captcha struct {
	Enabled         bool    `yaml:"enabled"`
	Length          int     `yaml:"length"`
	LifetimeSeconds int     `yaml:"lifetime_seconds"`
	Honeypot        bool    `yaml:"honeypot"`
	Lockout         Lockout `yaml:"lockout"`
}

type Lockout struct {
	Failures int `yaml:"failures"`
	Seconds  int `yaml:"seconds"`
}

type Limit struct {
	Requests   int `yaml:"requests"`
	PerSeconds int `yaml:"per_seconds"`
}

type RateLimits struct {
	Guests        Limit `yaml:"guests"`
	Users         Limit `yaml:"users"`
	AuthAttempts  Limit `yaml:"auth_attempts"`
	Registrations Limit `yaml:"registrations"`
}

type Registration struct {
	Enabled       bool   `yaml:"enabled"`
	ClosedMessage string `yaml:"closed_message"`
}

type About struct {
	Enabled bool   `yaml:"enabled"`
	Title   string `yaml:"title"`
	Body    string `yaml:"body"`
}

type Guests struct {
	Notes  int `yaml:"notes"`
	Boards int `yaml:"boards"`
}

type Admin struct {
	Usernames []string `yaml:"usernames"`
}

type Cooldowns struct {
	PostSeconds   int `yaml:"post_seconds"`
	NoteSeconds   int `yaml:"note_seconds"`
	ThreadsPerDay int `yaml:"threads_per_day"`
}

type Accounts struct {
	DeletionGraceHours int `yaml:"deletion_grace_hours"`
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
	ListenAddr     string   `yaml:"listen_addr"`
	CookieSecure   bool     `yaml:"cookie_secure"`
	ProxyHeader    string   `yaml:"proxy_header"`
	TrustedProxies []string `yaml:"trusted_proxies"`
}

type Database struct {
	URL string `yaml:"url"`
}

type Pagination struct {
	Notes   int `yaml:"notes"`
	Replies int `yaml:"replies"`
	Threads int `yaml:"threads"`
	Posts   int `yaml:"posts"`
	Archive int `yaml:"archive"`
	Users   int `yaml:"users"`
	Tags    int `yaml:"tags"`
	Search  int `yaml:"search"`
}

type BoardRules struct {
	BumpLimit      int    `yaml:"bump_limit"`
	PostLimit      int    `yaml:"post_limit"`
	MaxThreads     int    `yaml:"max_threads"`
	PreviewReplies int    `yaml:"preview_replies"`
	AnonName       string `yaml:"anon_name"`
}

type RulesOverride struct {
	BumpLimit      *int    `yaml:"bump_limit"`
	PostLimit      *int    `yaml:"post_limit"`
	MaxThreads     *int    `yaml:"max_threads"`
	PreviewReplies *int    `yaml:"preview_replies"`
	AnonName       *string `yaml:"anon_name"`
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
	if o.AnonName != nil {
		r.AnonName = *o.AnonName
	}
	return r
}

func (c Config) Textboard() bool { return c.Mode == ModeTextboard }

func Defaults() Config {
	return Config{
		Mode:       ModeCafe,
		Server:     Server{ListenAddr: ":8080"},
		Database:   Database{URL: "postgres://maidcafe:maidcafe@localhost:5432/maidcafe?sslmode=disable"},
		Pagination: Pagination{Notes: 30, Replies: 50, Threads: 10, Posts: 50, Archive: 50, Users: 30, Tags: 60, Search: 20},
		Tags:       Tags{MaxLength: 8, MaxPerPost: 3},
		Sidebar:    Sidebar{Boards: 2, Users: 5, Tags: 12},
		Accounts:   Accounts{DeletionGraceHours: 24},
		Cooldowns:  Cooldowns{PostSeconds: 20, NoteSeconds: 10, ThreadsPerDay: 10},
		Guests:     Guests{Notes: 10, Boards: 1},
		Captcha: Captcha{
			Enabled:         true,
			Length:          6,
			LifetimeSeconds: 300,
			Honeypot:        true,
			Lockout:         Lockout{Failures: 3, Seconds: 900},
		},
		RateLimits: RateLimits{
			Guests:        Limit{Requests: 60, PerSeconds: 60},
			Users:         Limit{Requests: 120, PerSeconds: 60},
			AuthAttempts:  Limit{Requests: 10, PerSeconds: 300},
			Registrations: Limit{Requests: 3, PerSeconds: 86400},
		},
		Registration: Registration{Enabled: true, ClosedMessage: "registration is closed right now. check back later."},
		About:        About{Enabled: true, Title: "about", Body: defaultAbout},
		Boards: Boards{Defaults: BoardRules{
			BumpLimit:      300,
			PostLimit:      1000,
			MaxThreads:     150,
			PreviewReplies: 5,
			AnonName:       "Anonymous",
		}},
	}
}

const defaultAbout = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

> Curabitur pretium tincidunt lacus. Nulla gravida orci a odio.

Nullam varius, turpis et commodo pharetra, est eros bibendum elit, nec luctus magna felis sollicitudin mauris. Integer in mauris eu nibh euismod gravida.`

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
	if c.Mode != ModeCafe && c.Mode != ModeTextboard {
		return fmt.Errorf("config: mode must be %q or %q", ModeCafe, ModeTextboard)
	}
	p := c.Pagination
	for name, v := range map[string]int{
		"pagination.notes": p.Notes, "pagination.replies": p.Replies, "pagination.threads": p.Threads,
		"pagination.posts": p.Posts, "pagination.archive": p.Archive, "pagination.users": p.Users,
		"pagination.tags": p.Tags, "pagination.search": p.Search,
	} {
		if v < 1 {
			return fmt.Errorf("config: %s must be at least 1", name)
		}
	}
	if c.Guests.Notes < 1 {
		return fmt.Errorf("config: guests.notes must be at least 1")
	}
	for name, v := range map[string]int{"sidebar.boards": c.Sidebar.Boards, "sidebar.users": c.Sidebar.Users, "sidebar.tags": c.Sidebar.Tags, "guests.boards": c.Guests.Boards} {
		if v < 0 {
			return fmt.Errorf("config: %s can't be negative", name)
		}
	}
	for name, v := range map[string]int{"cooldowns.post_seconds": c.Cooldowns.PostSeconds, "cooldowns.note_seconds": c.Cooldowns.NoteSeconds, "cooldowns.threads_per_day": c.Cooldowns.ThreadsPerDay} {
		if v < 0 {
			return fmt.Errorf("config: %s can't be negative (0 disables it)", name)
		}
	}
	cp := c.Captcha
	switch {
	case cp.Length < 4 || cp.Length > 10:
		return fmt.Errorf("config: captcha.length must be between 4 and 10")
	case cp.LifetimeSeconds < 30 || cp.LifetimeSeconds > 3600:
		return fmt.Errorf("config: captcha.lifetime_seconds must be between 30 and 3600")
	case cp.Lockout.Failures < 1:
		return fmt.Errorf("config: captcha.lockout.failures must be at least 1")
	case cp.Lockout.Seconds < 0:
		return fmt.Errorf("config: captcha.lockout.seconds can't be negative (0 disables it)")
	}
	rl := c.RateLimits
	for name, l := range map[string]Limit{
		"rate_limits.guests": rl.Guests, "rate_limits.users": rl.Users,
		"rate_limits.auth_attempts": rl.AuthAttempts, "rate_limits.registrations": rl.Registrations,
	} {
		if l.Requests < 0 || l.PerSeconds < 0 {
			return fmt.Errorf("config: %s can't be negative (0 disables it)", name)
		}
		if l.Requests > 0 && l.PerSeconds == 0 {
			return fmt.Errorf("config: %s.per_seconds must be set when requests is", name)
		}
	}
	if c.About.Enabled && (c.About.Title == "" || len(c.About.Title) > 40) {
		return fmt.Errorf("config: about.title must be 1-40 characters")
	}
	if c.Accounts.DeletionGraceHours < 0 {
		return fmt.Errorf("config: accounts.deletion_grace_hours can't be negative")
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
	case r.AnonName == "" || utf8.RuneCountInString(r.AnonName) > 30:
		return fmt.Errorf("config: %s.anon_name must be 1-30 characters", where)
	}
	return nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
