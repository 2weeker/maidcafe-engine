# Roadmap

## Known limits

- **Top sort is all-time.** Old popular notes stay on top forever. A time window (today / week)
  belongs in `config.yaml`.
- **No editing or deleting** notes or posts. Futaba-style self-delete (leaving a "post deleted"
  placeholder so numbering and `>>N` stay intact) is the planned approach.
- **No notifications** for likes, reposts, replies or `>>N` mentions.
- **No rate limiting** on registration or posting.
- **No moderation:** no roles, reports, bans, or capcodes.
- **Boards live in the database,** seeded by the first migration; there's no admin UI or config
  for them.
- **Text limits are constants** (note 300, post 4000, title 120, bio 500, art 24×80). Tag
  length and tags-per-post are configurable.
- **Tags can't be edited or removed** from a post, and tags themselves can't be renamed or
  deleted (no moderation yet).
- **Search has no stemming** (`brew` doesn't match `brewing`) and doesn't search profile art.
- **Following has no notifications,** and the following feed doesn't include replies.
- **Chats** have no per-message deleting, blocking, muting, kicking, moderators, channel topics
  or read receipts, and messages are stored unencrypted (until the retention sweep removes them). The live stream's hub is in-process, so multiple app instances would need
  Postgres `LISTEN/NOTIFY` (or similar) between them.
- **CSRF tokens live in process memory,** so running several app instances needs shared storage.
- **Expired sessions aren't swept** (they're ignored, just not deleted).
- **Formatting renders on every view.** Fine at this size; cache rendered HTML if busy pages slow down.
- **AA fonts aren't bundled.** Without MS PGothic or a compatible font installed, wide AA can
  drift. Bundling the free Saitamaar font in `web/static/fonts` would fix it.
- **Kaomoji with `*`** like `(*´∀｀*)` lose their asterisks to italics unless escaped (`\*`).
  Option: only treat `*` as emphasis when it's next to word characters.
- **Leading spaces in notes collapse** (standard markdown), which breaks small doodles outside a
  code fence.
- **`>>N` races:** in busy threads, someone else's reply can take the number you were about to
  reference. An edit window or a "your post will be #N" hint would help.
- **Profile art width** counts characters, so double-width characters can exceed 80 visual columns.
- **The `/random/` override in `config.yaml`** uses showcase limits (bump 10, posts 20, 6
  threads); relax it for normal use.

## Ideas from 2channel / futaba not yet built

| Feature | Notes |
| --- | --- |
| Catalog view | A grid of all threads with OP excerpts and reply counts |
| Daily per-thread IDs | `ID:aB3xK9pQ` hashes of user + day + board; follow people within a thread without names |
| Anonymous boards | A per-board setting to hide usernames, with a board-specific default name |
| Tripcodes | `name#secret` → `name ◆Xy9…` on anonymous boards |
| Capcodes | `## Admin` styling for staff posts (needs roles) |
| Post ranges in URLs | `/t/42/l50`, `/t/42/100-200`, `/t/42/5` |
| "New since last visit" | An unread marker per thread, "3 new" on the board list |
| Hide threads/posts | A per-user `[hide]` |
| Greentext styling | `>text` (no space) rendered in muted italic |
| Spoilers | `[spoiler]…[/spoiler]` revealed on hover/focus |
| Dice and fortunes | `!dice 2d6`, `!omikuji`, rolled when the post is saved |
