# Routes

All POSTs need a valid `_csrf` form field (every form includes it). "Login" means the route
redirects anonymous visitors to `/login`. Successful POSTs redirect (303) back to a page; failed
validation re-renders the form with the input kept.

## Pages

| Method | Path | Login | Page |
| --- | --- | --- | --- |
| GET | `/` | | Home: note composer with tag pickers, `recent`/`top`/`following` notes, boards + users + tags sidebar. `?sort=`, `?page=N` |
| GET | `/n/:id` | | A note: its parent (if a reply), the note, reply box, replies. `?page=N` |
| GET | `/u/:name` | | Profile: art, bio, follow button, counts, notes + reposts + replies. `?sort=top`, `?page=N` |
| GET | `/u/:name/followers`, `/u/:name/following` | | Follower / following lists. `?page=N` |
| GET | `/users` | | All users (`newest`, `name`) or the people you follow (`followed`). `?sort=`, `?page=N` |
| GET | `/tags` | | All tags + create form. `?sort=popular|newest|name`, `?page=N` |
| GET | `/tags/:name` | | A tag's notes (`?sort=`) or threads (`?show=threads`). `?page=N` |
| GET | `/search` | | `?q=`, `?in=all|notes|threads|users|tags`, `?page=N` |
| GET | `/dm` | yes | Inbox: your conversations, newest activity first. `?page=N` |
| GET | `/dm/channels` | yes | Public channels with join buttons. `?page=N` |
| GET | `/dm/c/:channel` | yes | A channel: log, member list, composer (members), or a join prompt. `?page=N` |
| GET | `/dm/c/:channel/events` | yes | Live stream for members (403 otherwise) |
| GET | `/dm/:name` | yes | Direct chat with a user (IRC-style). `?page=N` for older messages |
| GET | `/dm/:name/events` | yes | Server-Sent Events stream of new messages. `?after=ID` or the `Last-Event-ID` header replays anything newer |
| GET | `/boards` | | Every board with live/archived thread counts, post count and last activity |
| GET | `/b/:slug` | | Board index: subject list, thread previews, new-thread form. `?page=N` |
| GET | `/b/:slug/archive` | | Archived threads of a board. `?page=N` |
| GET | `/t/:id` | | Thread. `?page=N`, `?quote=N` prefills the reply box with `>>N` |
| GET | `/login`, `/register` | | Auth forms |
| GET | `/settings` | yes | Edit bio and text art |
| GET | `/static/*` | | CSS and fonts (embedded) |

## Actions

| Method | Path | Login | Fields | Effect |
| --- | --- | --- | --- | --- |
| POST | `/register` | | `username`, `password`, `password_confirm` | Create account, log in, go to profile |
| POST | `/login` | | `username`, `password` | Start a session |
| POST | `/logout` | | | End the session |
| POST | `/theme` | | `theme` (`light`/`eyesaver`/`dark`), `next` | Set the theme cookie, return to `next` |
| POST | `/notes` | yes | `body`, `tag1`…`tagN` | Post a note with up to N tags |
| POST | `/u/:name/follow` | yes | `on` (`1`/`0`), `next` | Follow or unfollow |
| POST | `/tags` | yes | `name` | Create a tag |
| POST | `/dm/:name` | yes | `body` | Send a direct message (creates the chat on first send; rejoins anyone who left). With `X-Requested-With: fetch`: 204 on success, 422/403 with the error as plain text |
| POST | `/dm/:name/leave` | yes | | Leave a direct chat |
| POST | `/dm/channels` | yes | `name`, `visibility` (`public`/`private`) | Create a channel |
| POST | `/dm/c/:channel` | yes | `body` | Send to a channel (members only) |
| POST | `/dm/c/:channel/join` | yes | | Join (public, or invited) |
| POST | `/dm/c/:channel/leave` | yes | | Leave |
| POST | `/dm/c/:channel/invite` | yes | `username` | Invite someone (members only) |
| POST | `/dm/c/:channel/decline` | yes | | Decline an invite |
| POST | `/n/:id/reply` | yes | `body` | Reply to a note |
| POST | `/n/:id/like` | yes | `on` (`1`/`0`), `next` | Like or unlike |
| POST | `/n/:id/repost` | yes | `on` (`1`/`0`), `next` | Repost or un-repost |
| POST | `/settings` | yes | `bio`, `art`, `action` (`preview`/`save`) | Preview or save the profile |
| POST | `/b/:slug` | yes | `title`, `body`, `tag1`…`tagN` | Start a thread with up to N tags |
| POST | `/t/:id` | yes | `body`, `sage` (`1`) | Reply in a thread |

## Status codes

| Code | When |
| --- | --- |
| 303 | After a successful POST |
| 204 | A message sent by the chat script |
| 401 | Wrong username or password |
| 403 | Missing or invalid CSRF token; sending to, streaming or inviting into a chat you're not in; joining a private channel uninvited |
| 404 | Unknown board, thread, note, user, tag, a conversation you're not part of, or a page number past the end |
| 409 | Replying to a full or archived thread |
| 422 | Validation failed (empty post, bad username, art too wide, bad or duplicate tag, unknown tag on a post, …) |
| 400 | Unknown theme, following yourself, or messaging yourself |

## Anchors

- `#pN`: post number N in a thread (`/t/42?page=2#p57`)
- `#tID`: a thread on a board index page (the subject list links to these)
- `#nID` / `#nID-r-USER`: a note, or a repost of it by USER, in a feed
- `#reply`: the reply box on thread and note pages
- `#latest`: the newest message in a DM conversation
