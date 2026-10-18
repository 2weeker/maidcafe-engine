# Users and following

## Following

Every profile has a `follow` button next to the name (`following`, solid black, once you do).
Clicking again unfollows. The button is a small form like the like button: it says which way it
goes, so a resubmitted form can't flip it, and it returns you to the same page.

- You can't follow yourself (the button isn't shown; a forged request gets a 400).
- Following someone twice is a no-op.
- Logged out, the button is a link to `/login`.

Under the name: `3 followers :: 5 following :: member since …`. Both counts link to lists:

| Page | Shows |
| --- | --- |
| `/u/:name/followers` | Who follows this user, newest first |
| `/u/:name/following` | Who this user follows, newest first |

Both are paginated (`pagination.users`) and show the same user cards as the users page.

## The following feed

The home page has a third sort tab, `following`, for logged-in users:

| Tab | Shows |
| --- | --- |
| `recent` | Everyone's top-level notes and reposts, by time |
| `top` | Everyone's top-level notes by likes |
| `following` | Top-level notes **and reposts** from people you follow, plus your own, by time |

Reposts count, as on Twitter: if someone you follow reposts a stranger's note, it appears with
"~x reposted". Logged out, the tab is hidden and `/?sort=following` shows a login prompt.
Following nobody yet shows a hint pointing to the users page.

## The users page (`/users`)

Members only: the users page, the follower and following lists and the users sidebar panel all
require a login ([accounts.md](accounts.md#logged-out-visitors)). Profiles stay public.

A grid of user cards, each with the name, follower and note counts, join date, the bio, and a
follow button. Sortable:

| Sort | Order |
| --- | --- |
| `newest` (default) | Most recent sign-ups first |
| `followed` | Only the people **you** follow, most recently followed first (logged in only) |
| `name` | Alphabetical |

Paginated by `pagination.users` (default 30).

## Sidebar

The home sidebar's **users** panel shows the five newest members with their join date and
follower count, and a `[all users]` link.

## Searching for users

`~name` or `@name` in the search bar jumps to that profile if it exists; otherwise it searches
usernames and bios. See [search.md](search.md).
