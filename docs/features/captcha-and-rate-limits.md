# Captcha, honeypot and rate limits

Ported from meth-engine's captcha (its `helper/captcha` packages) and adapted to the login and
registration forms. Everything is server-side, image-only and JavaScript-free, and every knob
lives in `config.yaml` ([configuration.md](../configuration.md)).

## Who sees what

| | Guests | Logged-in users |
| --- | --- | --- |
| `/login` and `/register` | the forms, with a captcha when `captcha.enabled` | not reachable: both redirect to your profile, so an account can't be used to skip the captcha and mint more accounts |
| Honeypot field | yes, when `captcha.honeypot` | n/a |
| Request rate limit | `rate_limits.guests`, per address | `rate_limits.users`, per account |
| Login / register attempt limit | `rate_limits.auth_attempts`, per address | same |
| New accounts per address | `rate_limits.registrations` | n/a |
| Posting cooldowns | n/a | `cooldowns` ([moderation.md](moderation.md#cooldowns)) |

## The captcha

```
captcha (type the letters and digits)
┌──────────────────────────┐
│  ~ 0 2 g c i y ~  . : .  │   192 x 64 PNG
└──────────────────────────┘
[                ]
```

- **What it is:** `captcha.length` characters (default 6) drawn from
  `abcdefghikmnoprstuvwxyz023456789`. The letters j, l and q and the digit 1 are left out, since
  they're too easy to confuse. Zeros are slashed so they can't be read as `o`.
- **How it's drawn:** the Go Medium font, a random size, a sine wave and a swirl through the
  middle, then 3-5 curved lines and 70-130 specks of noise. Black ink on a transparent
  background. The page gives it the paper color and inverts it in the dark theme.
- **Where it lives:** in memory, up to 50,000 at once (the oldest are dropped first). Nothing goes
  in the database. A restart forgets all open captchas and lockouts.
- **Bound to the form and the address:** a login captcha can't be used on the registration form,
  and only the address it was issued to can fetch its image or answer it. IPv4 addresses are
  matched exactly; IPv6 addresses are grouped by their /64.
- **Single use:** checking an answer burns the captcha, right or wrong. A failed form comes back
  with a fresh one.
- **Expiry:** `captcha.lifetime_seconds` (default 300). An expired captcha is refused and its
  image returns 404.
- **Answers:** case and surrounding spaces don't matter. Comparison is constant-time.
- **Image URL:** `/captcha/<32 hex>.png`, cached privately for five minutes.

### Lockout

After `captcha.lockout.failures` wrong answers (default 3) from one address on one form, that
form is locked for `captcha.lockout.seconds` (default 900) for that address. The form then shows
*!! too many wrong captchas. try again in about 15 minutes.* with no captcha and a disabled
submit button, and a POST gets `429` with a `Retry-After` header. Failures older than an hour are
forgotten, and a correct answer clears the count. Login and registration are counted
separately.

### Turning it off

`captcha.enabled: false` removes the image and the check entirely. Nothing is generated and
nothing is stored. The honeypot is a separate switch.

## The honeypot

Both forms carry an extra text field named `website`, labelled *leave this field empty*. It is
moved 10,000px off-screen rather than hidden with `display: none`, because form-filling bots
know to skip hidden fields. It has `tabindex="-1"` and `aria-hidden`, so keyboard and screen
reader users never reach it. A submission with anything in it is refused with a vague
*that didn't go through* (400), and the attempt is logged.

## Rate limits

All limits are token buckets kept in memory. `requests` is the burst, and the bucket refills
evenly over `per_seconds`. So `60` per `60` allows 60 requests at once, then one more per second.
Setting `requests: 0` turns a limit off.

| Key | Counts | Keyed by | Over the limit |
| --- | --- | --- | --- |
| `guests` | every page, form post and captcha image a logged-out visitor requests | address | `429` page: *slow down: too many requests. try again in N seconds.* |
| `users` | the same, for a logged-in user | account | the same |
| `auth_attempts` | every POST to `/login` and `/register`, whatever the outcome | address | the form again with a wait message, `429` |
| `registrations` | accounts successfully created | address | *too many new accounts from your address. try again in about N hours.*, `429` |

Static files (`/static/…`) are never counted. Every `429` carries `Retry-After`.

Logged-in users are counted per account, not per address, so people sharing an address (a
household, a campus) don't use up each other's allowance once they're logged in.

### Behind a reverse proxy

Limits and captchas key on the client address. Behind Caddy, nginx or a load balancer, every
request would otherwise appear to come from the proxy. Set:

```yaml
server:
  proxy_header: "X-Forwarded-For"
  trusted_proxies: ["10.0.0.5"]
```

The header is only believed when the request comes from a trusted proxy. With
`trusted_proxies` empty, loopback and private ranges are trusted, which suits a proxy on the same
host or Docker network. Leave `proxy_header` empty when the app faces the internet directly.

## Registration on and off

```yaml
registration:
  enabled: false
  closed_message: "registration is closed right now. check back later."
```

When it's off:

- `/register` shows the closed message with a link to log in.
- `POST /register` is refused with `403` and creates nothing.
- `[register]` disappears from the nav, the login page and every *log in or register* prompt.
- Existing accounts log in as usual. Admins listed in `config.yaml` are unaffected.

## Logged-in users and the auth forms

A logged-in user who opens `/login` or `/register` is sent to their own profile, and a POST to
either is refused the same way. Logging out first is the only way to the forms, and from there
the captcha, honeypot and limits apply like for any guest.

## Not covered

- Captchas and limits live in one process's memory. Several app instances behind one load
  balancer would each keep their own.
- The captcha has no audio or text alternative, same as meth-engine. Turning it off and relying
  on the honeypot and rate limits is the accessible fallback.
