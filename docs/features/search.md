# Search

The bar under the masthead on every page is a plain GET form: `/search?q=…&in=…`. No JavaScript.

```
[search://][ q                       ][ * all ▾ ][ go >> ]
```

## Scopes

| `in=` | Dropdown shows | Searches | Results |
| --- | --- | --- | --- |
| `all` (default) | `* all` | everything below | up to 5 of each kind, each group with a `[more …]` link |
| `notes` | `✎ notes` | note bodies | note cards, paginated |
| `threads` | `>> threads` | post bodies and thread titles | matching threads (first page), then matching posts linking to `#N` in their thread |
| `users` | `~ users` | usernames and bios | user cards |
| `tags` | `# tags` | tag names | tag chips with use counts |

Paginated scopes use `pagination.search` (default 20). A page past the last result is a 404,
like every other list.

Logged-out visitors get only `all`, `notes` and `threads`; `in=users` or `in=tags` falls back to
`all`, and `all` leaves out users and tags for them. `#tag` still jumps to the tag page, which
then asks them to log in.

## Shortcuts

| Type | Does |
| --- | --- |
| `#go` | Jumps to `/tags/go` if the tag exists; otherwise searches tag names for `go` |
| `~alice` or `@alice` | Jumps to `/u/alice` if the user exists; otherwise searches users for `alice` |

## Query syntax

Note and post bodies use Postgres full-text search (`websearch_to_tsquery`), so:

- `coffee tea` finds posts containing both words
- `"iced coffee"` finds the exact phrase
- `coffee -decaf` excludes a word
- `coffee or tea` finds either

Matching is on whole words, without stemming (`brew` doesn't match `brewing`), and is
case-insensitive. Results are ordered by relevance, then recency. Usernames, bios, tag names
and thread titles use substring matching instead, so `ali` finds `alice`.

Queries are trimmed and capped at 200 characters. Unbalanced quotes or stray operators are
tolerated (they just don't match anything).

## Indexes

Migration `005` adds GIN indexes on `notes.body`, `posts.body` and `threads.title`
(`to_tsvector('simple', …)`), and a pattern index on usernames. They're what keeps body search
fast as the site grows.
