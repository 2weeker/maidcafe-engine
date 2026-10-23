# Accounts

This page describes the default `cafe` mode. In [textboard mode](textboard-mode.md) registration
is off, profiles don't exist, and accounts are created by admins for moderators only.

## Registration (no email)

`/register` asks for three things: a username, a password, and the password again. Nothing else:
no email, no phone. Logged-out visitors also answer an image captcha, and the form carries a
hidden honeypot field ([captcha-and-rate-limits.md](captcha-and-rate-limits.md)). The form says it
plainly: *lose the password, lose the account*. There is no recovery.

Registration can be closed from `config.yaml` (`registration.enabled: false`); the register
links then disappear and `/register` shows a configurable message. New accounts per address are
capped by `rate_limits.registrations`.

| Field | Rule |
| --- | --- |
| Username | 3-20 characters, `a-z`, `0-9`, `_`. Case-insensitive: `Alice_1` is stored and shown as `alice_1`, and nobody else can take `ALICE_1` |
| Password | 8-72 characters (72 is bcrypt's limit) |
| Confirm | Must match |

Errors (*username is taken*, *password must be 8-72 characters*, *passwords do not match*)
re-render the form with the username kept. On success you're logged in and taken to your profile.

## Logging in and out

- `/login`: username (any case), password, and the captcha when enabled. A wrong combination
  says *wrong username or password* without revealing which part was wrong. Attempts per address
  are capped by `rate_limits.auth_attempts`.
- `[logout]` in the nav is a POST button; it deletes the session server-side and clears the cookie.

## Sessions

- A random token in the `mc_session` cookie (`HttpOnly`, `SameSite=Lax`, `Secure` when
  configured), valid for 30 days.
- The database stores only a SHA-256 of the token.
- Pages that need an account (posting, liking, settings) redirect to `/login` when you're not
  logged in.

## Logged-out visitors

The site is readable without an account, but the front page is a teaser and the people-and-tags
side of it is members-only:

| Guests see | Guests don't see |
| --- | --- |
| The first `guests.notes` notes on the home feed (default 10), `recent` or `top`, with a login prompt instead of older pages | Older notes on the home feed, the `following` tab |
| The first `guests.boards` boards in the sidebar (default 1), with a login prompt if there are more | The users and tags panels on the home page |
| `/boards`, every board, thread and archive, every note page, every profile | `/users`, `/tags`, `/tags/:name`, follower and following lists (all redirect to `/login`) |
| Search in `all`, `notes` and `threads` scopes | The `users` and `tags` search scopes; `all` doesn't include users or tags either |

The counts live in `config.yaml` under `guests` ([configuration.md](../configuration.md#guests)).

## Roles and bans

Accounts are `user` by default. Admins are named in `config.yaml`; they appoint mods. A banned
account can log in and read but not post, and sees the reason in a banner. See
[moderation.md](moderation.md).

## Deleting your account

At the bottom of `/settings`, under **delete account**: type your password and press
*schedule deletion*. Nothing is removed immediately.

- **Grace window:** `accounts.deletion_grace_hours` in `config.yaml` (default 24). Until it
  passes, a banner on every page says when the deletion will happen with a `[keep my account]`
  button, and the settings page offers the same. Cancelling clears the schedule entirely.
- **When the window passes,** a sweep (every minute) deletes the account: the user row,
  sessions, follows, likes, reposts, and every note and note reply they wrote (replies under
  their notes go with them). Tags they created stay, with no creator. The username becomes
  available again.
- **Board posts and threads are anonymized, not removed:** they show as `~deleted` with the body
  `[deleted]`, so other people's replies, thread numbering and `>>N` references keep working.
  Anonymized posts are left out of search.
- A wrong password leaves the account untouched and says so.

## Where your account shows up

- `[~yourname]` and `[settings]` in the nav
- `/u/yourname`: your profile ([profiles.md](profiles.md))
- `~yourname` on every note and post you write, linking to your profile
