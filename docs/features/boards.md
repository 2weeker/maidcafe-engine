# Boards (the textboard)

A 2channel/futaba-style forum: **boards** hold **threads**, threads hold numbered **posts**.

Seeded boards:

| Board | Description |
| --- | --- |
| `/meta/` | About this place |
| `/general/` | Anything goes. Mostly |
| `/tech/` | Computers, code, hardware |
| `/random/` | The void stares back (configured with tight limits for the showcase) |

`/boards` lists every board with its live and archived thread counts, total posts, last
activity and an archive link. The home sidebar shows only the first `sidebar.boards` of them
(default 2) with a `more boards >>` button.

## Board index (`/b/:slug`)

Top to bottom:

1. **Header:** name, description, and the board's rules: `bump limit 300 :: post limit 1000 ::
   150 live threads :: [archive]`.
2. **`[ start a new thread ]`:** a collapsible form (title ≤ 120 chars, body ≤ 4000) with tag
   dropdowns (see [tags.md](tags.md)). It opens by itself if your last attempt had an error.
3. **Subject list:** 2ch-style numbered list of the threads on this page with their post counts,
   linking to each preview below.
4. **Thread previews (futaba style):** each thread's title, status tag and `#tag` chips, its OP, then
   *"N replies omitted. [view thread]"*, then its latest replies (`preview_replies`, default 5),
   indented.
5. **Pagination:** `pagination.threads` threads per page (default 10).

Threads are ordered by **bump time**: a thread jumps to the top when someone replies, unless the
reply is sage or the thread is past its bump limit.

## Threads (`/t/:id`)

- Posts are numbered per thread: `#1` is the OP (tagged `OP`), then `#2`, `#3`…
- The header shows the post count and any status tag: `bump limit`, `full`, `archived`.
- 50 posts per page (`pagination.posts`), with `prev`/`next` above and below.
- After you reply, you land on your new post, on whichever page it is.

### `>>N` references and backlinks

- Writing `>>12` in a post links to post #12 of the same thread, on the right page.
- Post #12 then lists **backlinks** underneath: `replies: >>15 >>31`. They're stored when the
  reply is posted, so a post on page 1 shows replies that live on page 3.
- `>>N` to a post that doesn't exist (yet) stays plain text.
- References inside `` `code` ``, code blocks and `[aa]` blocks are ignored.
- **Cross-board:** `>>>/tech/` links to a board; `>>>/tech/42` links to thread 42. Thread IDs are
  site-wide, so the number is a thread, not a post.

### Quoting a post

Every post has a `[>>]` button. It opens the thread's last page with `>>N` already in the reply
box: the no-JS version of 2ch's click-the-number-to-quote. (`/t/:id?quote=N`.)

### Editing and deleting

Your own posts have `[edit]` and `[delete]` in their header. Edits get an `(edited)` marker.
Deleting leaves `[deleted]` in place so numbering and references survive. Moderators see the same
tools on every post (`[remove]` / `[restore]`) plus `[lock]`, `[pin]` and `[delete thread]` on the
thread; see [moderation.md](moderation.md).

### Sage

The reply form has a **sage (don't bump)** checkbox. A sage reply is posted normally but leaves
the thread where it is on the board. Sage posts carry a dashed `sage` tag.

## Thread lifecycle

All three limits are per board, set in [`config.yaml`](../configuration.md).

| Limit | Default | What happens |
| --- | --- | --- |
| `bump_limit` | 300 posts | Replies past this no longer bump the thread. It's tagged `bump limit` and slowly sinks |
| `post_limit` | 1000 posts | The thread is **full**: tagged `full`, and the reply box is replaced by *"this thread has reached N posts. start a new one on /board/."* |
| `max_threads` | 150 | Creating a thread beyond this archives the least-recently-bumped live threads |

Moderators can also **lock** a thread (no replies, tagged `locked`) and **pin** it (sorted first).
Posting is rate-limited by the `cooldowns` settings ([moderation.md](moderation.md#cooldowns)).

If a reply is refused because the thread filled up (or was archived) while you were typing, the
page shows your text in a read-only box so you can copy it.

### Archive (`/b/:slug/archive`)

Archived threads leave the board index and appear in the archive list (title, author, post
count, archive time), newest first. They stay readable at their usual URL, tagged `archived`,
with no reply box.

## Formatting

Board posts use the same markdown subset as notes ([formatting.md](formatting.md)), including
`>greentext` and `[spoiler]` tags, plus `>>N` references and `[aa]` Shift-JIS art blocks.

## Limits

| | |
| --- | --- |
| Thread title | 1-120 characters |
| Post body | 1-4000 characters |
