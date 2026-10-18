# Moderation and the admin panel

## Roles

| Role | Can |
| --- | --- |
| `user` | Everything described in the other feature docs |
| `mod` | All of that, plus: remove and restore posts, remove notes, edit anyone's post or note, lock, pin and delete threads, ban and unban users, read the mod log |
| `admin` | All of that, plus: appoint and demote mods, ban mods, manage boards, view the loaded config |

**Admins are set in `config.yaml`:**

```yaml
admin:
  usernames:
    - testr
```

The list is authoritative. On every start the named accounts get `admin` (and any ban on them is
lifted), and any account holding `admin` that isn't listed is demoted to `user`. Accounts that
don't exist yet become admin the next time the app starts after they register. Admins can't be
banned or demoted from the panel; edit the file and `make restart`.

Mods are appointed by admins from `/admin/users`.

## Where moderation happens

- **`[admin]`** appears in the nav for mods and admins and opens the panel.
- **Inline tools** show up for mods where they're needed:
  - on a thread page: `[lock]` / `[unlock]`, `[pin]` / `[unpin]`, `[delete thread]`
  - on every post and note: `[edit]` and `[remove]` (`[restore]` on a removed post)
  - on a profile: `moderate` (jumps to the user in the panel) or `banned` (lifts the ban)

## The panel (`/admin`)

| Page | Who | What |
| --- | --- | --- |
| dashboard | mods | Counts (users, mods, banned, pending deletion, notes, threads, posts, tags) and the ten latest mod actions |
| users | mods | Search by username; see role, activity, ban and deletion status; ban with a reason / unban. Admins also see a role toggle (`user` ↔ `mod`) |
| boards | admins | Edit each board's name, description and position inline; delete empty boards; create new ones |
| log | mods | Every moderation action, newest first, paginated |
| config | admins | The configuration the running app loaded, as YAML, with the database password hidden |

Every form is a plain POST; no JavaScript anywhere.

## What each action does

| Action | Effect |
| --- | --- |
| **Remove post** | The body becomes `[removed by moderator]`; the post stays in the thread so numbering and `>>N` links keep working. `[restore]` brings it back |
| **Remove note** | The note and the replies under it are deleted |
| **Edit** someone's post or note | Saved with an `(edited)` marker, and logged |
| **Lock thread** | No more replies (409 to anyone who tries); tagged `locked`. Unlock reverses it |
| **Pin thread** | Sorted to the top of its board index; tagged `pinned` |
| **Delete thread** | Hard delete of the thread and all its posts, after a confirmation page |
| **Ban** | The user can still log in and read, but every posting route (notes, replies, threads, tags, likes, reposts, follows) is refused with 403 and the reason. A banner tells them. Mods can ban users; only admins can ban mods; nobody can ban admins or themselves |
| **Role** | Admins switch a user between `user` and `mod` |
| **Boards** | Create (slug `a-z0-9`, 1-16 chars; name; description; position), edit, or delete a board with no threads |

## The mod log

`mod_actions` records who did what to which target, with a detail string (the reason for a ban,
the post's author, the thread title). Deleting and editing your **own** content isn't logged;
doing it to someone else's is.

## Self-service editing and deletion

Any user can edit or delete their own posts and notes:

- `[edit]` opens a page with the text; saving marks the item `(edited)`.
- `[delete]` goes to a confirmation page first (no JavaScript dialogs). A deleted board post
  stays as `[deleted]`; a deleted note is removed with its replies.
- Deleted posts can't be edited or quoted, and don't appear in search.

## Cooldowns

```yaml
cooldowns:
  post_seconds: 20
  note_seconds: 10
  threads_per_day: 10
```

| Key | Limits |
| --- | --- |
| `post_seconds` | Time between a user's board posts (replies and new threads) |
| `note_seconds` | Time between a user's notes and note replies |
| `threads_per_day` | Threads a user can start in any rolling 24 hours |

Exceeding one returns `429` with a message like *slow down: you can post again in 12 seconds*.
`0` disables that limit. Mods and admins are subject to the same limits.
