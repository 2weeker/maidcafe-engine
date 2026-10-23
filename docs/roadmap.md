# Roadmap

## Known limits

- **Top sort is all-time.** Old popular notes stay on top forever. A time window (today / week)
  belongs in `config.yaml`.
- **Moderation is manual:** there's no report button, no word filters, no IP-level tooling,
  and no appeal flow for bans. Futaba-style self-delete (leaving a "post deleted"
  placeholder so numbering and `>>N` stay intact) is the planned approach.
- **No notifications** for likes, reposts, replies or `>>N` mentions.
- **Captchas, lockouts and rate-limit counters are in memory.** A restart clears them, and
  several app instances wouldn't share them.
- **The captcha has no audio or text alternative.** Turning it off keeps the honeypot and rate
  limits as the fallback.
- **Text limits are constants** (note 300, post 4000, title 120, bio 500, art 24×80). Tag
  length and tags-per-post are configurable.
- **Tags can't be edited or removed** from a post, and tags themselves can't be renamed or
  deleted (no moderation yet).
- **Search has no stemming** (`brew` doesn't match `brewing`) and doesn't search profile art.
- **Following has no notifications,** and the following feed doesn't include replies.
- **No private messaging.** A DM/chat feature was built and then removed to keep the site
  JavaScript-free; it's in git history (`git log -- app/controllers/dm.go`) if ever wanted back.
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
| Per-board anonymity | `mode: textboard` makes the whole site anonymous ([textboard-mode.md](features/textboard-mode.md)); a per-board switch, and a board-specific default name, aren't built |
| Secure tripcodes | Classic `name#secret` → `name ◆Xy9…` is built ([tripcodes.md](features/tripcodes.md)); a server-salted `##secure` variant isn't |
| Capcodes | `## Admin` styling for staff posts (needs roles) |
| Post ranges in URLs | `/t/42/l50`, `/t/42/100-200`, `/t/42/5` |
| "New since last visit" | An unread marker per thread, "3 new" on the board list |
| Hide threads/posts | A per-user `[hide]` |
| Dice and fortunes | `!dice 2d6`, `!omikuji`, rolled when the post is saved |
