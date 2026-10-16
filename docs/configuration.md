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
- any `sidebar` count or `chat.retention_hours` is negative,
- any page size is below 1,
- `bump_limit` < 1, `post_limit` < 2, `max_threads` < 1, or `preview_replies` < 0,
- `bump_limit` is greater than `post_limit`,

for the defaults and for every board override.

## Reference

```yaml
server:
  listen_addr: ":8080"
  cookie_secure: false

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
  messages: 50

tags:
  max_length: 8
  max_per_post: 3

sidebar:
  boards: 2
  users: 5
  tags: 12

chat:
  retention_hours: 24

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
| `messages` | 50 | Messages per page in a DM conversation (the inbox uses `users`) |

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

### `chat`

| Key | Default | Meaning |
| --- | --- | --- |
| `retention_hours` | 24 | Delete chat messages (DMs and channels, including join/leave lines) older than this. `0` keeps them forever. A sweep runs every minute |

### `boards.defaults`

Rules for every board unless overridden.

| Key | Default | Meaning |
| --- | --- | --- |
| `bump_limit` | 300 | Once a thread has this many posts, replies stop moving it to the top of the board |
| `post_limit` | 1000 | Once a thread has this many posts it is **full**: the reply box is replaced by a notice |
| `max_threads` | 150 | Live threads per board. Creating a thread beyond this archives the least-recently-bumped ones |
| `preview_replies` | 5 | Latest replies shown under each OP on the board index (0 = OP only) |

"Posts" always includes the opening post: a thread with an OP and 9 replies has 10 posts.

### `boards.overrides`

A map from board slug to any subset of the keys above. Unlisted keys fall back to
`boards.defaults`. Slugs that don't match a board are harmless (and ignored).

The shipped `config.yaml` gives `/random/` deliberately tight limits so the lifecycle
features are easy to see (see [showcase.md](showcase.md)). Remove or relax that override for
normal use.

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
  migration. Add or rename boards with SQL for now.
- Text limits (note 300 chars, post 4000, title 120, bio 500, art 24×80) are constants in the
  models.
- Themes are a fixed list in `web/templates/page.go`.

See [roadmap.md](roadmap.md).
