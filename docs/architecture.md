# Architecture

## Principles

- **No JavaScript, with one exception.** Every interaction is a link or an HTML form. Toggles
  (likes, reposts, theme) are tiny POST forms that redirect back to where you were. The one
  exception is the real-time DM chat (`web/static/js/chat.js`), which enhances a page that still
  works as a plain form without it.
- **Server-rendered.** templ compiles views to Go; Fiber serves them. One binary with the CSS and
  fonts embedded.
- **Minimal dependencies.** Fiber, templ, pgx, bcrypt, goldmark, a YAML parser. No ORM, no
  frontend toolchain.
- **Postgres does the bookkeeping.** Counters, post numbers and limits are enforced inside
  transactions so concurrent requests can't corrupt them.

## Layout (MVC)

```
cmd/server/main.go         entrypoint: config, database, middleware, routes, listen
app/
  config/                  config.yaml loading, env overrides, validation
  database/                pgx pool, migration runner
    migrations/*.sql       schema, applied in filename order, embedded in the binary
  models/                  M: all SQL lives here (Store methods), validation, domain types
  controllers/             C: one handler per route, error page, page/pager helpers
  middleware/              session lookup (LoadUser), RequireUser guard
  realtime/                in-process pub/sub hub for live DM streams
  routes/                  the URL table
web/
  templates/               V: *.templ views + small Go helpers (URLs, view structs)
  markup/                  user text -> safe HTML (restricted markdown, >>N, [aa])
  static/css/style.css     all styling
  static/fonts/            Inter variable font (regular + italic woff2)
  web.go                   embeds static/ into the binary
```

`Inter/` at the repository root is the original font download (variable TTFs, static cuts, the OFL
license). Only the two `.woff2` files under `web/static/fonts/` are served; `Inter/` is excluded
from the Docker build.

### Dependencies between layers

```
routes -> controllers -> models  -> database
              |            ^
              v            |
          templates ------/   (views read model types)
              |
              v
           markup
```

Controllers call the model `Store`, then render a templ component. Views import models (for
types) and config (for board rules) but never query the database.

## A request, end to end

1. **Fiber middleware**, in order: `recover` (turn panics into 500s), `logger`, static files under
   `/static/*` (served from the embedded filesystem with a 1 hour cache), `LoadUser` (resolve the
   session cookie into a user), `csrf` (validate POSTs).
2. **Route** dispatches to a controller method. Routes that change data for a user are wrapped
   in `RequireUser`, which redirects anonymous visitors to `/login`.
3. **Controller** reads params and form values, calls `Store` methods, and either redirects
   (after a successful POST, the Post/Redirect/Get pattern) or renders a page.
4. **Validation errors** (empty post, bad username…) re-render the same page with a `422` or
   `409` and the user's input preserved, plus an `!! message` banner.
5. **Unhandled errors** go to `controllers.ErrorHandler`, which renders a styled error page:
   `models.ErrNotFound` → 404, Fiber errors keep their code, anything else is logged and becomes
   a 500 with no details leaked.

Every page is built from `templates.Page` (title, current user, CSRF token, theme, current path
for the theme button, optional error), which `controllers.page()` fills.

## Rendering

- Views are `.templ` files compiled by `templ generate` into `*_templ.go` (gitignored, generated
  during the Docker build, or locally with `make generate`).
- templ escapes all interpolated text. The only raw HTML inserted comes from `web/markup`, which
  produces HTML from user text with raw HTML disabled (see
  [features/formatting.md](features/formatting.md)).
- ASCII art (banner, footer, profile art) is passed as Go strings into `<pre>` so templ's
  whitespace handling can't disturb it.

## Security model

| Concern | How it's handled |
| --- | --- |
| Passwords | bcrypt (default cost). Length 8-72 bytes. Unknown usernames still run a bcrypt comparison so login timing doesn't reveal which accounts exist |
| Sessions | 32 random bytes, base64url, in an `HttpOnly`, `SameSite=Lax` cookie `mc_session`, valid 30 days. Only the SHA-256 of the token is stored, so a database leak doesn't leak live sessions. Logout deletes the row |
| CSRF | Fiber's CSRF middleware on every POST. Each form carries a hidden `_csrf` field; requests without a valid token get `403` |
| XSS | templ escapes everything; the markdown renderer has raw HTML parsing removed, so `<script>` is shown as text; `javascript:` and similar URLs are emptied |
| Links | User links get `rel="nofollow ugc noopener noreferrer"` |
| Open redirects | The `next` field used by the theme and like/repost forms must be a local path (`/…`, not `//…` or `/\…`); anything else redirects to `/` |
| Request size | Bodies are capped at 64 KB |
| Cookies over HTTPS | Set `server.cookie_secure: true` (or `COOKIE_SECURE=true`) |

**Not yet:** rate limiting, account lockout, moderation tools, CSRF storage shared across
multiple app instances (tokens live in memory, which is fine for one instance). See
[roadmap.md](roadmap.md).

## Configuration flow

`config.Load()` → `Config` is passed to `controllers.New` and to the middleware setup in
`main.go`. Controllers ask `ctl.Config.Boards.For(slug)` for a board's rules and pass the numbers
down to the models (`CreateThread`'s `maxThreads`, `CreateReply`'s `ReplyOptions`), so models stay
free of config types.

## Build and deploy

- **Dockerfile:** stage one (`golang:1.27-alpine`) downloads modules, runs `go tool templ generate`
  (templ is pinned as a Go tool in `go.mod`), and builds a static binary with `CGO_ENABLED=0`.
  Stage two is `alpine:3.22` with an unprivileged `maidcafe` user.
- **Compose:** `db` (postgres:18-alpine with a health check and a named volume) and `app`
  (depends on a healthy db, mounts `config.yaml`, exposes `APP_PORT`).
- **Shutdown:** SIGINT/SIGTERM close the realtime hub (ending open DM streams), then trigger
  Fiber's graceful shutdown with a 10-second limit.
