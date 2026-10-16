# Testing

## Unit tests

```sh
make generate      # templ output must exist for the packages to compile
go test ./...
```

| Package | Covers |
| --- | --- |
| `app/config` | Missing/empty/partial files keep defaults, per-board overrides merge onto defaults, env vars beat the file, unknown keys and invalid limits (including tag settings) are rejected with the offending key named |
| `app/models` | `CleanArt` (indentation kept, trailing space and outer blank lines trimmed, tabs, CRLF, control chars, 24-line and 80-column limits, error names the line) and `CleanBio` |
| `app/realtime` | Pair keys are order-independent, publishing reaches only that pair's subscribers, a full subscriber never blocks the sender, unsubscribing cleans up, closing is idempotent |
| `web/markup` | Inline formatting, raw HTML shown as text, quote-needs-a-space, no lazy continuation, nested quotes, code fences untouched, hard line breaks, no headings, no indented code, `javascript:` stripped, `rel` on links, images become links, lists; `>>N` links/dead refs/refs in quotes, inline code and fences; cross-board links; `PostRefs`; `[aa]` blocks (including backticks inside and unclosed blocks) |

The model and controller code that talks to Postgres is covered by end-to-end runs rather than
unit tests (below).

## End-to-end checks

Each feature was verified against a throwaway Compose stack (`docker compose -p mctest`, port
8090, deleted afterwards) with curl and headless Firefox screenshots:

- registration, duplicate usernames, login/logout, CSRF rejection (403), anonymous POSTs
  redirected to `/login`
- notes: posting, validation (422), escaping, likes/reposts/replies with idempotent counters,
  `recent`/`top` ordering, pagination with the sort preserved, 404 past the last page
- themes: the light → eyesaver → dark cycle, the legacy `coffee` cookie, rejecting unknown themes,
  off-site `next` values ignored
- profiles: preview without saving, art validation messages, art cleanup, escaping, scaling on
  narrow screens, the column ruler lining up with the textarea
- boards, with a config of bump 3 / post 5 / 2 threads / 2 previews / 3 posts per page: gapless
  post numbers, backlinks across pages, dead refs, refs in code ignored, bump limit stops
  bumping, the post limit refuses replies (409) and keeps the unsent text, sage doesn't bump,
  archiving the oldest thread and refusing replies to it, previews with omitted counts,
  `?quote=N`, `[aa]` rendering
- follows, tags and search: creating tags (length, character, duplicate and anonymous
  refusals), tagged notes and threads, unknown-tag refusal with input kept, follow/unfollow
  idempotence, self-follow 400, follower counts and lists, the following feed for followers,
  non-followers and anonymous visitors, users page sorts, sidebar panels, tag pages (notes and
  threads tabs), `#tag` and `~user` search shortcuts, every search scope, odd query strings
- DM live chat (with curl as the client): event-stream headers, instant delivery to the other
  participant, `mine` styling on your own lines, isolation between conversations,
  `Last-Event-ID` replay after a disconnect, unread badge staying at zero while watching,
  `fetch` sends returning 204/422, and the no-JS form path. The browser script itself was
  checked by reading and screenshots, not by an automated browser run
- migrations applied cleanly to the existing local database each time

## Simulated community

[showcase.md](showcase.md) describes the population seeded by four agents through a small HTTP
client that drives the site exactly like a browser (real forms, real CSRF tokens). It doubles as
a broad smoke test; problems the agents ran into are listed there.
