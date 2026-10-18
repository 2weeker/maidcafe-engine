# Routes

All POSTs need a valid `_csrf` form field (every form includes it). "Login" means the route
redirects anonymous visitors to `/login`; "guest" means the route is for logged-out visitors only
and sends logged-in users to their profile. Successful POSTs redirect (303) back to a page; failed
validation re-renders the form with the input kept.

## Pages

| Method | Path | Login | Page |
| --- | --- | --- | --- |
| GET | `/` | | Home: note composer with tag pickers, `recent`/`top`/`following` notes, boards + users + tags sidebar. `?sort=`, `?page=N`. Logged out: `guests.notes` notes, `guests.boards` boards, no users/tags panels, no paging |
| GET | `/n/:id` | | A note: its parent (if a reply), the note, reply box, replies. `?page=N` |
| GET | `/u/:name` | | Profile: art, bio, follow button, counts, notes + reposts + replies. `?sort=top`, `?page=N` |
| GET | `/u/:name/followers`, `/u/:name/following` | yes | Follower / following lists. `?page=N` |
| GET | `/users` | yes | All users (`newest`, `name`) or the people you follow (`followed`). `?sort=`, `?page=N` |
| GET | `/tags` | yes | All tags + create form. `?sort=popular|newest|name`, `?page=N` |
| GET | `/tags/:name` | yes | A tag's notes (`?sort=`) or threads (`?show=threads`). `?page=N` |
| GET | `/search` | | `?q=`, `?in=all|notes|threads|users|tags`, `?page=N`. Logged out: only `all`, `notes`, `threads`, and `all` skips users and tags |
| GET | `/boards` | | Every board with live/archived thread counts, post count and last activity |
| GET | `/b/:slug` | | Board index: subject list, thread previews, new-thread form. `?page=N` |
| GET | `/b/:slug/archive` | | Archived threads of a board. `?page=N` |
| GET | `/t/:id` | | Thread. `?page=N`, `?quote=N` prefills the reply box with `>>N` |
| GET | `/login`, `/register` | guest | Auth forms, with a captcha when enabled. Logged-in users are redirected to their profile. `/register` shows the closed message when registration is off |
| GET | `/captcha/:id.png` | | A captcha image; only for the address it was issued to, 404 once used or expired |
| GET | `/about` | | The about page from `config.yaml`; 404 when disabled |
| GET | `/settings` | yes | Edit bio and text art |
| GET | `/t/:id/p/:no/edit`, `/n/:id/edit` | yes | Edit your own post / note (mods: anyone's) |
| GET | `/t/:id/p/:no/delete`, `/n/:id/delete`, `/t/:id/delete` | yes | Confirmation pages for deleting a post / note / thread (thread: mods) |
| GET | `/admin` | mod | Dashboard |
| GET | `/admin/users` | mod | User management. `?q=`, `?page=N` |
| GET | `/admin/log` | mod | Mod log. `?page=N` |
| GET | `/admin/boards` | admin | Board management |
| GET | `/admin/config` | admin | The loaded configuration |
| GET | `/static/*` | | CSS and fonts (embedded) |

## Actions

| Method | Path | Login | Fields | Effect |
| --- | --- | --- | --- | --- |
| POST | `/register` | guest | `username`, `password`, `password_confirm`, `captcha_id`, `captcha_value`, `website` (must be empty) | Create account, log in, go to profile. `403` when registration is closed, `422` wrong captcha, `429` locked out or rate limited |
| POST | `/login` | guest | `username`, `password`, `captcha_id`, `captcha_value`, `website` (must be empty) | Start a session. `422` wrong captcha, `429` locked out or rate limited |
| POST | `/logout` | | | End the session |
| POST | `/theme` | | `theme` (`light`/`eyesaver`/`dark`), `next` | Set the theme cookie, return to `next` |
| POST | `/notes` | yes | `body`, `tag1`…`tagN` | Post a note with up to N tags |
| POST | `/u/:name/follow` | yes | `on` (`1`/`0`), `next` | Follow or unfollow |
| POST | `/tags` | yes | `name` | Create a tag |
| POST | `/n/:id/reply` | yes | `body` | Reply to a note |
| POST | `/n/:id/like` | yes | `on` (`1`/`0`), `next` | Like or unlike |
| POST | `/n/:id/repost` | yes | `on` (`1`/`0`), `next` | Repost or un-repost |
| POST | `/settings` | yes | `bio`, `art`, `action` (`preview`/`save`) | Preview or save the profile |
| POST | `/settings/delete` | yes | `password` | Schedule account deletion after the grace window |
| POST | `/settings/undelete` | yes | `next` | Cancel a scheduled deletion |
| POST | `/b/:slug` | yes | `title`, `body`, `tag1`…`tagN` | Start a thread with up to N tags |
| POST | `/t/:id` | yes | `body`, `sage` (`1`) | Reply in a thread |
| POST | `/t/:id/p/:no/edit`, `/n/:id/edit` | yes | `body` | Save an edit |
| POST | `/t/:id/p/:no/delete`, `/n/:id/delete` | yes | | Delete (post → `[deleted]`, note → removed) |
| POST | `/t/:id/p/:no/restore` | mod | | Restore a removed post |
| POST | `/t/:id/lock`, `/t/:id/pin` | mod | `on` (`1`/`0`) | Lock/unlock, pin/unpin |
| POST | `/t/:id/delete` | mod | | Delete a thread |
| POST | `/u/:name/ban`, `/u/:name/unban` | mod | `reason`, `next` | Ban / unban |
| POST | `/u/:name/role` | admin | `role` (`user`/`mod`), `q` | Change role |
| POST | `/admin/boards` | admin | `slug`, `name`, `description`, `position` | Create a board |
| POST | `/admin/boards/:slug` | admin | `name`, `description`, `position`, `action` (`delete`) | Edit or delete a board |

## Status codes

| Code | When |
| --- | --- |
| 303 | After a successful POST |
| 401 | Wrong username or password |
| 403 | Missing or invalid CSRF token; banned account on a posting route; editing what isn't yours; mod/admin pages without the role |
| 409 | Replying to a full, archived or locked thread; deleting a board that has threads |
| 410 | Editing a deleted post |
| 429 | A cooldown hasn't passed yet |
| 404 | Unknown board, thread, note, user, tag, or a page number past the end |
| 422 | Validation failed (empty post, bad username, art too wide, bad or duplicate tag, unknown tag on a post, wrong password on delete, …) |
| 400 | Unknown theme, or following yourself |

## Anchors

- `#pN`: post number N in a thread (`/t/42?page=2#p57`)
- `#tID`: a thread on a board index page (the subject list links to these)
- `#nID` / `#nID-r-USER`: a note, or a repost of it by USER, in a feed
- `#reply`: the reply box on thread and note pages
