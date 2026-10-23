# Textboard mode

maidcafe runs in one of two modes, chosen in `config.yaml`:

```yaml
mode: cafe        # the default: microblog + boards, open registration
mode: textboard   # anonymous boards only, accounts for staff
```

Changing it needs a restart (`make restart`). The database is the same in both modes, so a site
can switch back and forth; what's hidden in one mode is still there in the other.

## What textboard mode does

| Area | cafe | textboard |
| --- | --- | --- |
| Front page | note composer, feed, sidebar | the boards list, with the search bar above it |
| Posting on boards | logged-in users | anyone, no account needed |
| Author shown on posts | `~username` linking to the profile | **Anonymous**; staff can sign a post (below) |
| Notes, profiles, following, users page, tags | on | off: every `/n/`, `/u/:name`, `/users` and `/tags` URL is a 404 |
| Tags on threads | tag dropdowns and `#tag` chips | hidden |
| Registration | open unless `registration.enabled: false` | off: `/register` is a 404 and there is no register link |
| Login | `[login]` in the nav | a `[staff]` link in the footer opens `/login` |
| Search | `all` / `notes` / `threads` / `users` / `tags` | thread titles and posts only; no scope dropdown |
| Names | the account's username | the board's default name, or a name with a classic tripcode set on `/name` ([tripcodes.md](tripcodes.md)) |
| Captcha | login and register | login, plus the thread and reply forms for posters without an account; the honeypot is on every form |
| Admin panel | as described in [moderation.md](moderation.md) | the same, minus the notes and tags tiles |
| Settings page | bio, text art, delete account | delete account only |

Everything else is unchanged: boards, threads, `>>N`, sage, archives, bump and post limits,
moderation tools, themes, the about page, the captcha on the login form, request rate limits.

## Anonymous posting

In textboard mode the new-thread form and the reply box are shown to everyone. A post made
without an account is stored with no author and renders under the board's default name
(`anon_name`, **Anonymous** by default), or as `name ◆tripcode` if the poster set one on `/name`
([tripcodes.md](tripcodes.md)). Nobody can edit or delete it later except a moderator
(`[remove]` / `[restore]`), since there is no owner to prove.

Both forms carry the image captcha for posters without an account and the honeypot field for
everyone, with the same lockout as the login form
([captcha-and-rate-limits.md](captcha-and-rate-limits.md)). Logged-in staff skip the captcha.

Cooldowns still apply, keyed by address instead of account (IPv4 exactly, IPv6 by /64):

| Key | For anonymous posters |
| --- | --- |
| `cooldowns.post_seconds` | Seconds between posts from one address |
| `cooldowns.threads_per_day` | Threads one address may start per day, refilling evenly across the day |

Over the limit returns `429` with the usual *slow down* message and a `Retry-After` header. These
counters are in memory, like the captcha and request limits.

## Staff accounts

Accounts exist only for moderators and admins. Since nobody can register, they're created two
ways:

**From the shell**, for the first admin:

```sh
make adduser NAME=alice            # asks for the password
make adduser NAME=bob ROLE=mod
PASSWORD=... make adduser NAME=alice
```

This runs `maidcafe adduser <name> [user|mod]` inside the app container. Names listed under
`admin.usernames` in `config.yaml` become admins immediately. Running the binary directly:
`PASSWORD=... ./maidcafe adduser alice`.

**From the admin panel**, once an admin exists: `/admin/users` has a `[ new account ]` form
(username, password, role `user` or `mod`), visible to admins. Creating an account is written to
the mod log.

Both ways work in cafe mode too.

### Signing a post

A logged-in mod or admin posts anonymously by default. The thread and reply forms show a
**sign as ~name** checkbox; ticking it attaches the account, and the post shows `~name` with a
green `## mod` or `## admin` capcode. Signed posts can be edited and deleted by their author like
any owned post.

### Logged-in accounts with the `user` role

If a `user`-role account exists (made before switching modes, or by an admin), it can log in,
but its posts are anonymous: only mods and admins can sign.

## What `[deleted]` means now

A post stored without an account was previously always a post from a deleted account, and
rendered as `[deleted]`. Now that anonymous posts also have no account, a post is a tombstone
only when its `deleted_at` is set: migration `011_anonymous_posts.sql` marks existing orphaned
posts that way, and the account purge marks a user's posts deleted before dropping the user.
Nothing changes visually in cafe mode.
