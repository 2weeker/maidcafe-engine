# Configuration

The app reads one YAML file at startup, then lets a few environment variables override it.
Changes need a restart: `make restart`.

## Where the file lives

| Situation | Path |
| --- | --- |
| Docker Compose (normal) | `./config.yaml` is mounted read-only at `/etc/maidcafe/config.yaml`, and `CONFIG_PATH` points there |
| Running the binary directly | `config.yaml` in the working directory, or whatever `CONFIG_PATH` says |

If the file doesn't exist, the app logs `config: … not found, using defaults` and runs with the
defaults below. An empty file also means defaults. A file that sets only some keys keeps the
defaults for the rest.

## Validation

Loading is strict. The app refuses to start, and says why, when:

- a key is unknown (catches typos like `bump_limt`),
- `tags.max_length` is outside 1-32 or `tags.max_per_post` outside 0-20,
- any `sidebar` count, `guests.boards`, `accounts.deletion_grace_hours` or a `cooldowns` value is
  negative,
- any page size or `guests.notes` is below 1,
- `bump_limit` < 1, `post_limit` < 2, `max_threads` < 1, or `preview_replies` < 0,
- `bump_limit` is greater than `post_limit`,
- `captcha.length` is outside 4-10, `captcha.lifetime_seconds` outside 30-3600,
  `captcha.lockout.failures` below 1 or `captcha.lockout.seconds` negative,
- a `rate_limits` value is negative, or `requests` is set without `per_seconds`,
- `about.enabled` is true and `about.title` isn't 1-40 characters,

for the defaults and for every board override.

## Reference

```yaml
server:
  listen_addr: ":8080"
  cookie_secure: false
  proxy_header: ""
  trusted_proxies: []

database:
  url: "postgres://maidcafe:maidcafe@localhost:5432/maidcafe?sslmode=disable"

pagination:
  notes: 30
  replies: 50
  threads: 10
  posts: 50
  archive: 50
  users: 30
  tags: 60
  search: 20

tags:
  max_length: 8
  max_per_post: 3

sidebar:
  boards: 2
  users: 5
  tags: 12

guests:
  notes: 10
  boards: 1

accounts:
  deletion_grace_hours: 24

captcha:
  enabled: true
  length: 6
  lifetime_seconds: 300
  honeypot: true
  lockout:
    failures: 3
    seconds: 900

rate_limits:
  guests:
    requests: 60
    per_seconds: 60
  users:
    requests: 120
    per_seconds: 60
  auth_attempts:
    requests: 10
    per_seconds: 300
  registrations:
    requests: 3
    per_seconds: 86400

registration:
  enabled: true
  closed_message: "registration is closed right now. check back later."

about:
  enabled: true
  title: "about"
  body: |
    Lorem ipsum dolor sit amet, ...

admin:
  usernames:
    - testr

cooldowns:
  post_seconds: 20
  note_seconds: 10
  threads_per_day: 10

boards:
  defaults:
    bump_limit: 300
    post_limit: 1000
    max_threads: 150
    preview_replies: 5
  overrides:
    random:
      bump_limit: 10
      post_limit: 20
      max_threads: 6
      preview_replies: 3
    meta:
      max_threads: 30
```

### `server`

| Key | Default | Meaning |
| --- | --- | --- |
| `listen_addr` | `":8080"` | Address the HTTP server binds to |
| `cookie_secure` | `false` | Mark session, CSRF and theme cookies `Secure`. **Set `true` when serving over HTTPS.** |
| `proxy_header` | `""` | Header holding the real client address behind a reverse proxy, such as `X-Forwarded-For`. Empty means use the connection's address |
| `trusted_proxies` | `[]` | Proxy addresses or CIDR ranges whose `proxy_header` is believed. Empty with a header set trusts loopback and private ranges |

### `database`

| Key | Default | Meaning |
| --- | --- | --- |
| `url` | `postgres://maidcafe:maidcafe@localhost:5432/maidcafe?sslmode=disable` | Postgres connection URL |

### `pagination`

| Key | Default | Where |
| --- | --- | --- |
| `notes` | 30 | Home timeline and profile note lists |
| `replies` | 50 | Replies on a note's page (`/n/:id`) |
| `threads` | 10 | Threads per board index page (each shows previews, so keep it small) |
| `posts` | 50 | Posts per thread page |
| `archive` | 50 | Threads per board archive page |
| `users` | 30 | Users per page on `/users` and follower/following lists |
| `tags` | 60 | Tags per page on `/tags` |
| `search` | 20 | Results per page for a single search scope |

### `tags`

| Key | Default | Meaning |
| --- | --- | --- |
| `max_length` | 8 | Longest allowed tag name, in characters (1-32) |
| `max_per_post` | 3 | Tag dropdowns shown on the note and thread composers, and the most tags a post can carry (0-20; 0 hides tagging) |

Lowering `max_length` later doesn't rename existing longer tags; they stay usable.

### `sidebar`

How many items the home page sidebar shows. Each panel links to the full page.

| Key | Default | Panel |
| --- | --- | --- |
| `boards` | 2 | First boards by position, then a `more boards >>` button to `/boards` |
| `users` | 5 | Newest members |
| `tags` | 12 | Most-used tags |

### `guests`

What a visitor who isn't logged in sees on the home page. See
[features/accounts.md](features/accounts.md#logged-out-visitors) for everything else guests
can and can't reach.

| Key | Default | Meaning |
| --- | --- | --- |
| `notes` | 10 | Notes on the home feed (`recent` or `top`), with no older pages; a login prompt stands in for the pager when more exist |
| `boards` | 1 | Boards in the sidebar; a login prompt replaces `more boards >>` when more exist |

The users and tags panels aren't shown to guests at all, and that isn't configurable.

### `accounts`

| Key | Default | Meaning |
| --- | --- | --- |
| `deletion_grace_hours` | 24 | How long a scheduled account deletion can be cancelled before the sweep removes the account. `0` deletes on the next sweep (within a minute) |

### `admin`

| Key | Default | Meaning |
| --- | --- | --- |
| `usernames` | `[]` | Accounts that are admins. Applied on every start: listed accounts become admin, unlisted admins are demoted. See [moderation.md](features/moderation.md) |

### `cooldowns`

| Key | Default | Meaning |
| --- | --- | --- |
| `post_seconds` | 20 | Minimum seconds between one user's board posts |
| `note_seconds` | 10 | Minimum seconds between one user's notes |
| `threads_per_day` | 10 | Threads one user may start per rolling 24 hours |

`0` disables a limit.

### `captcha`

Applies to the login and registration forms, for logged-out visitors only. See
[captcha-and-rate-limits.md](features/captcha-and-rate-limits.md).

| Key | Default | Meaning |
| --- | --- | --- |
| `enabled` | `true` | Show and check the image captcha |
| `length` | 6 | Characters per captcha (4-10) |
| `lifetime_seconds` | 300 | How long a captcha stays answerable (30-3600) |
| `honeypot` | `true` | Add the hidden `website` field and refuse forms that fill it. Independent of `enabled` |
| `lockout.failures` | 3 | Wrong answers from one address on one form before it's locked |
| `lockout.seconds` | 900 | How long the lockout lasts. `0` never locks |

### `rate_limits`

Each limit is `requests` allowed at once, refilling evenly over `per_seconds`. `requests: 0`
turns it off.

| Key | Default | Counts |
| --- | --- | --- |
| `guests` | 60 / 60s | Requests from a logged-out visitor, per address |
| `users` | 120 / 60s | Requests from a logged-in user, per account |
| `auth_attempts` | 10 / 300s | Login and register submissions, per address |
| `registrations` | 3 / 86400s | Accounts created, per address |

### `registration`

| Key | Default | Meaning |
| --- | --- | --- |
| `enabled` | `true` | `false` closes sign-ups: the register links disappear and `/register` shows `closed_message` |
| `closed_message` | *registration is closed right now. check back later.* | Shown on `/register` while closed |

### `about`

| Key | Default | Meaning |
| --- | --- | --- |
| `enabled` | `true` | Serve `/about` and link it from the footer |
| `title` | `about` | Heading, tab title and footer link label (1-40 characters) |
| `body` | lorem ipsum | The page text, in the site's [formatting](features/formatting.md) |

## Environment variables

These win over the file. Docker Compose sets them.

| Variable | Overrides | Notes |
| --- | --- | --- |
| `CONFIG_PATH` | — | Path of the YAML file (default `config.yaml`) |
| `DATABASE_URL` | `database.url` | Compose builds it from the Postgres variables below |
| `LISTEN_ADDR` | `server.listen_addr` | |
| `COOKIE_SECURE` | `server.cookie_secure` | `true` or anything else for false |

Compose-level variables (read from your shell or a `.env` file):

| Variable | Default | Used for |
| --- | --- | --- |
| `APP_PORT` | 8080 | Host port mapped to the app |
| `POSTGRES_USER` | maidcafe | Database user |
| `POSTGRES_PASSWORD` | maidcafe | Database password |
| `POSTGRES_DB` | maidcafe | Database name |
| `COOKIE_SECURE` | false | Passed through to the app |

## What isn't configurable yet

- The list of boards (slug, name, description, order) lives in the database, seeded by the first
  migration and managed from the admin panel (`/admin/boards`).
- Text limits (note 300 chars, post 4000, title 120, bio 500, art 24×80) are constants in the
  models.
- Themes are a fixed list in `web/templates/page.go`.

See [roadmap.md](roadmap.md).
