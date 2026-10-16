# Tags

Tags are short labels that can be attached to **notes** and **threads**. They are created on
their own page, then picked from a dropdown when composing; nobody types a tag into a post.

## Creating tags (`/tags`)

Logged-in users see a `#` input and a `create tag` button at the top of the tags page.

| Rule | Default | Where it's set |
| --- | --- | --- |
| Length | 1-8 characters | `tags.max_length` in `config.yaml` |
| Characters | `a-z 0-9 - _` | fixed |
| Case | stored lowercase (`#Tea` becomes `#tea`) | fixed |
| Uniqueness | one tag per name; *that tag already exists* otherwise | fixed |

Errors re-render the page with the typed name kept. A leading `#` is ignored. On success you
land on the new tag's page.

The tags page lists every tag with its note count, thread count, creator and creation time,
sortable by `popular` (default), `newest` or `name`, paginated by `pagination.tags`.

## Tagging a note or thread

The note composer on the home page and the new-thread form on every board end with a row of
dropdowns:

```
tags  [ -- ▾ ] [ -- ▾ ] [ -- ▾ ]          [ post note ]
```

- One dropdown per allowed slot: `tags.max_per_post` (default 3). Set it to 0 to hide the
  pickers entirely.
- Each dropdown lists every existing tag, alphabetically, with `--` for "none".
- Picking the same tag twice counts once.
- If no tags exist yet the row says *no tags yet. create one.*
- A tag name that doesn't exist (only possible by editing the form) is refused with *unknown tag*.

Tags go on the **thread**, not on individual replies, so a tag page lists whole threads. Note
replies can't be tagged either; only top-level notes can.

Tags can't be changed after posting, because posts can't be edited yet.

## Where tags show

- As `#name` chips under a note's body, and next to a thread's title on board indexes and
  thread pages. Each chip links to the tag's page.
- The home sidebar's **tags** panel: the 12 most-used tags, larger and bolder the more they're
  used, with a `[all tags]` link.

## Tag pages (`/tags/:name`)

```
#go
2 notes :: 1 thread :: made by ~alice
[notes] [threads]
[recent] [top]
...
```

- **notes**: every top-level note with the tag, sortable `recent`/`top`, paginated.
- **threads**: every thread with the tag, across all boards, by bump time; archived threads are
  marked.

## Searching for tags

`#go` in the search bar jumps straight to `/tags/go` if it exists; otherwise it searches tag
names containing `go`. See [search.md](search.md).
