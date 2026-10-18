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
| `app/captcha` | Ported from meth-engine: ids are 32 hex characters and answers use the reduced alphabet; single use, wrong answers burn the captcha, case and spaces ignored; binding to form and address; images only for the issuing address; expiry and sweeping; capacity drops the oldest; deterministic PNGs of the right size; configurable length and lifetime |
| `app/captcha/captchaimg` | Glyphs stay on the canvas, seeds distort differently, noise is added but doesn't drown the text, zeros are slashed |
| `app/limits` | IPv4 and IPv4-mapped keys, IPv6 grouped by /64; token buckets allow a burst, refill, cap at the burst and keep keys apart; `requests: 0` disables; `Ready` doesn't spend; failure lockouts lock, expire, reset on success and forget old failures |
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
- migrations applied cleanly to the existing local database each time
- captcha and limits, on a throwaway stack with tight limits and four config variants: captcha
  image served as a 192×64 PNG and 404 for unknown ids; wrong answer 422 with a fresh captcha;
  lockout 429 with `Retry-After`, locked form with no captcha and a disabled button; login and
  register locked separately; honeypot 400 and logged; no captcha for a logged-in user; the
  third account from one address refused; login attempts capped; guest and user request bursts
  hitting 429; registration closed (links gone, closed message, POST 403 and no row written);
  custom about title and markdown body; about disabled 404 and footer link gone

## Simulated community

[showcase.md](showcase.md) describes the population seeded by four agents through a small HTTP
client that drives the site exactly like a browser (real forms, real CSRF tokens). It doubles as
a broad smoke test; problems the agents ran into are listed there.
