# Accounts

## Registration (no email)

`/register` asks for three things: a username, a password, and the password again. Nothing else:
no email, no phone, no captcha (yet). The form says it plainly: *lose the password, lose the
account*. There is no recovery.

| Field | Rule |
| --- | --- |
| Username | 3-20 characters, `a-z`, `0-9`, `_`. Case-insensitive: `Alice_1` is stored and shown as `alice_1`, and nobody else can take `ALICE_1` |
| Password | 8-72 characters (72 is bcrypt's limit) |
| Confirm | Must match |

Errors (*username is taken*, *password must be 8-72 characters*, *passwords do not match*)
re-render the form with the username kept. On success you're logged in and taken to your profile.

## Logging in and out

- `/login`: username (any case) and password. A wrong combination says *wrong username or
  password* without revealing which part was wrong.
- `[logout]` in the nav is a POST button; it deletes the session server-side and clears the cookie.

## Sessions

- A random token in the `mc_session` cookie (`HttpOnly`, `SameSite=Lax`, `Secure` when
  configured), valid for 30 days.
- The database stores only a SHA-256 of the token.
- Pages that need an account (posting, liking, settings) redirect to `/login` when you're not
  logged in. Anonymous visitors can read everything.

## Where your account shows up

- `[~yourname]` and `[settings]` in the nav
- `/u/yourname`: your profile ([profiles.md](profiles.md))
- `~yourname` on every note and post you write, linking to your profile
