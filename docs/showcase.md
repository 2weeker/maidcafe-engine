# Showcase: the simulated community

The local instance (<http://localhost:8080>) was populated by four AI agents, each playing two
residents. They used a small HTTP client that drives the site like a browser: real forms, real
CSRF tokens, one cookie jar per resident. Every feature was exercised through the public UI; no
data was written to the database directly.

## Residents

All passwords are `maidcafe-demo` (except `testr`, the original hand-made test account).

| User | Persona | Notes | Posts | Profile |
| --- | --- | --- | --- | --- |
| `mochi_mod` | Friendly, organized owner of /meta/ | 10 | 9 | Bio + 13-line teacup art |
| `latte_art` | Quiet text artist | 10 | 6 | Bio + 21×60 art: box-drawing frame, ░▒▓█ shading, ⣿ braille rosette |
| `nanashi_42` | Deadpan 2ch veteran, lives on `>>N` and sage | 6 | 31 | Bio |
| `kaomoji_k` | Excitable AA fan | 5 | 29 | Bio + kaomoji banner art |
| `kernel_kat` | Terse systems programmer | 7 | 59 | Bio + 12-line terminal and keyboard art |
| `byte_bard` | Enthusiastic hobbyist | 7 | 59 | Bio |
| `sleepy_sora` | Night-shift café worker | 20 | 0 | Bio |
| `night_owl_nao` | Witty social glue | 22 | 1 | Bio + owl art |
| `testr` | The original tester | 15 | 4 | Bio + art |

Themes were set per resident to exercise the cookie (eyesaver: mochi_mod, kaomoji_k, sleepy_sora;
dark: latte_art, nanashi_42, kernel_kat, night_owl_nao). Themes are per browser, so you'll still
see your own.

## By the numbers

| | |
| --- | --- |
| Top-level notes | 59 (home feed runs to 2 pages) |
| Note replies | 43 (three chains 3 levels deep, rooted at notes 87, 88, 155) |
| Likes / reposts | 99 / 15 |
| Threads | 28 (2 archived) |
| Board posts | 198 (19 sage, 11 with `[aa]` art) |
| `>>N` backlinks | 126 |

| Board | Live threads | Archived | Posts |
| --- | --- | --- | --- |
| /meta/ | 2 | 0 | 14 |
| /general/ | 12 | 0 | 95 |
| /tech/ | 6 | 0 | 33 |
| /random/ | 6 | 2 | 56 |

## Where to look, feature by feature

| Feature | Where |
| --- | --- |
| Email-free accounts | Any profile, e.g. `/u/mochi_mod`; log in as any resident |
| Notes with markdown | `/`: lists (106), `> quote` + reply (98, 102, 117), strike (106), inline code (135), links (116), fenced ASCII doodle (96), braille (107), box drawing (136) |
| Likes, reposts, `top` sort | `/?sort=top`: notes 102, 98, 96, 87 have 5 likes each; 96 has 3 reposts |
| Repost entries in the feed | `/`: "~x reposted" lines; note 96 appears once per reposter |
| Note replies and chains | `/n/87` (4 replies), `/n/102` → 113 → 115 |
| Feed pagination | `/?page=2` |
| Profile bio + text art | `/u/latte_art` (the big one), `/u/kernel_kat`, `/u/mochi_mod`, `/u/night_owl_nao`, `/u/kaomoji_k` |
| `>>N` + backlinks | `/t/54` (feature requests: #4 shows `replies: >>6 >>8 >>9`), `/t/52` |
| Backlinks across pages | `/t/56`: posts on page 2 reference page 1; #1 shows backlinks into page 2 |
| Thread pagination | `/t/56` (65 posts, 2 pages) |
| Board index pagination | `/b/general` (12 live threads → `?page=2`) |
| Previews + omitted counts | `/b/random`, `/b/general` |
| 2ch subject list | Top of every board index |
| Sage | `/t/45` posts #6, #12, #13; tags marked `sage` across /tech/ and /random/ |
| Bump limit | `/t/44` "AA appreciation thread" (tagged `bump limit`) |
| Post limit (full) | `/t/45` "count to 20, no talking": 20 posts, reply box replaced by the notice |
| Archive | `/b/random/archive`: threads 45 and 51; both read-only |
| `[aa]` Shift-JIS art | `/t/44` (Mona, Giko, orz, a maid serving tea), `/t/56` #2, `/t/52` #3 |
| Code blocks, refs inside code stay literal | `/t/49` (#3), `/t/46`, `/t/50` |
| Raw HTML shown as text | `/t/49` #1 and #4 (the `<b>` joke) |
| Cross-board links | `/t/52` #1 (`>>>/random/`, `>>>/tech/`), `/t/44` #8, `/t/54` |
| Quote button `[>>]` | On every post |
| Search mockup, sidebar placeholders | Every page / the home sidebar |

## What the residents talked about

- **/meta/:** the welcome thread with house rules and a formatting primer; a feature-requests
  thread asking for real search, the users/tags panels, an edit window, and a "you're replying
  as #N" hint (plus a struck-out request for a jukebox).
- **/random/:** an AA appreciation thread, the counting thread (which only reached 17 because
  of sage chatter, then filled up), "what did you eat today", greentext stories, "rate the thread
  above", kaomoji-only replies, small unpopular opinions.
- **/tech/:** editor wars, homelabs, Go vs Rust for a weekend project, "ask the stack anything"
  about this site, small daily tools.
- **/general/:** the 65-post "what are you listening to" thread, tea ratings, tiny wins, lurkers
  saying hi, desks, books, snack tier lists, cozy games, weekend plans, plants, today-I-learned.
- **Notes:** café shift stories, decaf hot takes, DNS jokes, doodles, and a lot of kaomoji.

## Issues the simulation surfaced

| Finding | Status |
| --- | --- |
| `[aa]` mid-sentence or inside inline code opened an art block | **Fixed**: `[aa]` must now start a line. Formatting happens on display, so the affected posts (`/t/44` #1, `/t/52` #1) render correctly now |
| `*` in kaomoji like `(*´∀｀*)` becomes italics | Standard markdown; escape as `\*`. Logged in the [roadmap](roadmap.md) |
| Leading spaces in plain notes collapse, breaking little doodles | Standard markdown; use a code fence or `[aa]`. Logged in the roadmap |
| `>>N` references can point at the wrong post when several people reply at once | Posts can't be edited yet; residents posted corrections in `/t/56` (#62, #63, #65). An edit window and a post-number hint are on the roadmap |
| A stray post "x" in `/t/47` #6 from a client misfire | There's no delete yet. #7 explains it in character ("my cat stepped on enter"). Remove with SQL if you like |
| A note reposted by two people appears twice in the recent feed | By design: each repost is its own feed entry, as on Twitter |
| A page past the last one is a full 404 page | By design |

The microblogger agent (sleepy_sora, night_owl_nao) was stopped early when the session was
wrapped up; everything it posted is in place.

## Resetting

To remove the showcase data and start fresh: `make clean && make up` (this deletes **all** data,
including `testr`). To keep the data but restore normal /random/ limits, edit the `random`
override in `config.yaml` and `make restart`.
