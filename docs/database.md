# Database

PostgreSQL 18. All SQL lives in `app/models` (queries) and `app/database/migrations` (schema).

## Migrations

`database.Migrate` runs at every startup:

1. creates `schema_migrations (version, applied_at)` if missing,
2. lists the embedded `migrations/*.sql` files in filename order,
3. applies each file not yet recorded, **in its own transaction**, then records it.

A failing migration rolls back and stops the app. Migrations are tracked by filename only, so
never rename one that has been applied; add a new numbered file instead.

| File | Adds |
| --- | --- |
| `001_init.sql` | `users`, `sessions`, `boards` (seeded with meta/general/tech/random), `threads`, `posts`, `notes` |
| `002_note_interactions.sql` | Note replies (`reply_to_id`), like/repost/reply counters, `note_likes`, `note_reposts` |
| `003_profile_art.sql` | `users.art` |
| `004_board_features.sql` | Per-thread post numbers (`posts.no`, backfilled), `posts.sage`, `threads.archived_at`, `post_refs` (backfilled from existing `>>N`) |
| `005_follows_tags_search.sql` | `follows`, `tags`, `note_tags`, `thread_tags`, full-text indexes for search |
| `012_tripcodes.sql` | `posts.name` and `posts.trip` for anonymous posts with a [tripcode](features/tripcodes.md) |
| `011_anonymous_posts.sql` | Marks posts that already had no author as deleted (`deleted_at`), so that from now on a post with no `user_id` and no `deleted_at` is an anonymous post ([textboard mode](features/textboard-mode.md)) |
| `010_moderation.sql` | `users.role/banned_at/ban_reason`, `posts.edited_at/deleted_at/removed`, `notes.edited_at`, `threads.locked_at/pinned`, `mod_actions` |
| `009_account_deletion.sql` | `users.delete_at`; `threads.user_id` and `posts.user_id` become nullable with `ON DELETE SET NULL` so a deleted author leaves `~deleted` posts behind |
| `008_drop_chat.sql` | Drops the chat tables (`conversations`, `chat_members`, `messages`) that migrations 006 and 007 created before the feature was removed; a no-op on fresh databases |

## Tables

### `users`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | bigserial | |
| `username` | text unique | `^[a-z0-9_]{3,20}$`, stored lowercase |
| `password_hash` | text | bcrypt |
| `bio` | text | ≤ 500 chars, markdown source |
| `art` | text | ≤ 2100 chars in the DB; the app enforces 24 lines × 80 columns |
| `created_at` | timestamptz | |
| `delete_at` | timestamptz | null, or when a scheduled deletion takes effect |
| `role` | text | `user`, `mod` or `admin` |
| `banned_at`, `ban_reason` | | set while banned |

Deleting a user cascades to sessions, follows, likes, reposts and notes; threads and posts keep
their rows with `user_id` set to null and render as `~deleted` / `[deleted]`.

### `sessions`

| Column | Notes |
| --- | --- |
| `token_hash` | SHA-256 of the cookie token (primary key) |
| `user_id` | cascades on user delete |
| `created_at`, `expires_at` | 30-day lifetime; expired rows are ignored (not yet swept) |

### `boards`

`id`, `slug` (`^[a-z0-9]{1,16}$`, unique), `name`, `description`, `position` (sort order).

### `threads`

| Column | Notes |
| --- | --- |
| `board_id`, `user_id` | owner board and OP author |
| `title` | 1-120 chars |
| `reply_count` | posts minus the OP; `posts = reply_count + 1` |
| `created_at`, `bumped_at` | the board index sorts live threads by `bumped_at` |
| `archived_at` | null while live; set when pushed out by `max_threads` |
| `locked_at` | set while a moderator has locked replies |
| `pinned` | sorted to the top of the board index |

### `posts`

| Column | Notes |
| --- | --- |
| `thread_id`, `user_id` | |
| `no` | position in the thread, 1 = OP; unique per thread; what `>>N` refers to |
| `body` | 1-4000 chars, markdown source |
| `sage` | the reply was posted without bumping |
| `edited_at` | set when the body is edited |
| `deleted_at`, `removed` | deleted by the author (`removed` false) or a moderator (`removed` true); the body renders as `[deleted]` / `[removed by moderator]` |
| `created_at` | |

### `mod_actions`

`actor_id` (null if the mod's account is gone), `action`, `target` (a path or `~user`),
`detail`, `created_at`. Append-only.

### `post_refs`

`(thread_id, to_no, from_no)`: post `from_no` contains `>>to_no`. Read as backlinks on `to_no`.
`CHECK (to_no < from_no)`: you can only reference earlier posts.

### `notes`

| Column | Notes |
| --- | --- |
| `user_id`, `body` | body 1-300 chars |
| `reply_to_id` | null for top-level notes; set for replies (cascades) |
| `like_count`, `repost_count`, `reply_count` | denormalized for sorting and display |
| `created_at` | |

### `note_likes`, `note_reposts`

`(user_id, note_id)` primary key, plus `created_at`. One like and one repost per user per note.

### `follows`

`(follower_id, followee_id)` primary key, `created_at`, `CHECK (follower_id <> followee_id)`.
Indexed both ways so follower lists and following feeds are cheap.

### `tags`

| Column | Notes |
| --- | --- |
| `name` | unique, `^[a-z0-9_-]{1,32}$`; the app enforces the configured `max_length` on creation |
| `created_by` | nullable, set null if the creator is deleted |
| `created_at` | |

### `note_tags`, `thread_tags`

`(note_id, tag_id)` and `(thread_id, tag_id)` primary keys, cascading on either side. Tags are
read back into each note/thread row as a `text[]` subquery (`array_agg`), so listing pages don't
need an extra query per item.

## Consistency rules

- **Counters move only when a row changes.** Liking inserts with `ON CONFLICT DO NOTHING` and
  bumps `like_count` only if a row was inserted; unliking only decrements if a row was deleted.
  Repeating a like or unliking something you never liked changes nothing, and counts can't go
  negative.
- **Post numbers are gapless.** A reply first runs one `UPDATE threads … RETURNING` that locks the
  thread row, checks it isn't archived or full, increments `reply_count`, and decides whether to
  bump. The returned count becomes the new post's `no`. Two simultaneous replies are serialized
  on that row lock.
- **Limits are exact.** The same `UPDATE` only matches while `reply_count + 1 < post_limit`. If it
  matches nothing, the app checks why and reports *not found*, *archived* or *full*.
- **Archiving happens with thread creation.** After inserting a thread, the same transaction
  archives every live thread on that board beyond the newest `max_threads` by bump order.
- **Feeds are computed, not stored.** The recent feed is a `UNION ALL` of top-level notes and
  repost rows ordered by time; the viewer's own like/repost state comes from `EXISTS` subqueries.

## Indexes worth knowing

- `threads_board_live_idx (board_id, bumped_at DESC, id DESC) WHERE archived_at IS NULL`: the board index
- `posts_thread_no_idx (thread_id, no)` unique: thread pages and number lookups
- `notes_toplevel_recent_idx` / `notes_toplevel_top_idx`: the two home feed sorts
- `note_reposts_recent_idx`: interleaving reposts into the feed
- `follows_follower_idx` / `follows_followee_idx`: following feed and follower lists
- `note_tags_tag_idx` / `thread_tags_tag_idx`: tag pages
- `notes_body_fts_idx`, `posts_body_fts_idx`, `threads_title_fts_idx` (GIN): full-text search
- `users_username_pattern_idx`: username prefix search
