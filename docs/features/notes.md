# Notes (the microblog)

Notes are short posts (≤ 300 characters) that behave like a Twitter-style feed. They live on the
home page and on profiles; boards and threads are a separate world ([boards.md](boards.md)).

## Posting

The box at the top of the home page (`what's on your mind, ~you?`). Notes support the
[formatting](formatting.md) subset: bold, italic, strike, code, links, lists, `> quotes`.
`>>N` stays literal in notes (there's no thread for it to point into), but `>>>/board/` links work.

## Every note's action bar

```
~author  re: ~parent                        2026-10-16 07:09   <- timestamp links to the note's page
note body
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
[ reply 3 ]  [ repost 1 ]  [ liked 5 ]
```

| Button | Does |
| --- | --- |
| `reply N` | Opens the note's page at the reply box |
| `repost N` / `reposted N` | Repost or undo. Solid black means you've reposted |
| `like N` / `liked N` | Like or undo. Solid black means you've liked |

Logged out, the buttons are links to `/login`.

**Without JavaScript:** like and repost are tiny forms. Each one states the change it wants
(`on=1` or `on=0`) instead of "toggle", so resubmitting a form (back button, double click) can
never flip it the wrong way. After the click you land back on the same page, sort and position
(the form carries `next=/current/path#this-note`).

## Replies

- A reply is a note with a parent. Post one from the parent's page (`/n/:id`).
- **The note page** shows the parent (if the note is itself a reply) joined by a connector line,
  the note in a heavier box, the reply box, then the replies, oldest first and paginated.
- **Replies stay out of the home feed** to keep it readable. They appear on the author's profile
  with a `re: ~parent` link.
- Chains can go as deep as people like; each level is its own page.

## Reposts

Reposting puts the note back at the top of the **recent** home feed with a context line:

```
~bob reposted
~alice                                      2026-10-16 07:09
the original note
```

It also appears on the reposter's profile. One repost per person per note; the note shows the
total.

## Tags

The composer ends with a row of tag dropdowns (one per allowed slot, default 3). Tags are chosen
from the ones created on `/tags`; they appear as `#name` chips under the note and link to the
tag's page. Replies can't be tagged. See [tags.md](tags.md).

## Sorting

Tabs on the home page and every profile:

| Sort | Home | Profile |
| --- | --- | --- |
| `recent` | Top-level notes and reposts, by time | The user's notes, replies and reposts, by time |
| `top` | Top-level notes by likes, newest first among ties; no repost entries | The user's notes and replies by likes |
| `following` | Notes and reposts from people you follow, plus your own, by time (logged in only) | — |

`top` is all-time for now. See [users-and-following.md](users-and-following.md) for following.

## Pagination

30 notes per page (configurable: `pagination.notes`). Pages are `?page=N` links:
`<< newer | [ page 2 ] | older >>` (`prev`/`next` under `top`). The sort carries through
(`/?sort=top&page=2`). The composer only shows on page 1. A page past the end is a 404.

## Limits

| | |
| --- | --- |
| Note / reply length | 1-300 characters (after trimming) |
| Likes | one per user per note |
| Reposts | one per user per note |
