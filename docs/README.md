# maidcafe-engine

A text-only, (almost) JavaScript-free social platform: part microblog, part 2channel/futaba-style
textboard. Accounts need no email. Pages are plain server-rendered HTML styled as black-and-white
web 1.0 with neo-brutalist touches.

**Stack:** Go 1.27 · Fiber v3 · templ · PostgreSQL 18 · goldmark · Docker Compose · Make

## Contents

| Doc | What's in it |
| --- | --- |
| [getting-started.md](getting-started.md) | Running it, `make` targets, resetting, logs, psql |
| [configuration.md](configuration.md) | Every `config.yaml` setting and environment variable |
| [architecture.md](architecture.md) | Project layout (MVC), request flow, security model |
| [database.md](database.md) | Schema, migrations, how counts and numbering stay consistent |
| [routes.md](routes.md) | Every URL and form endpoint |
| [features/accounts.md](features/accounts.md) | Email-free registration, sessions, logout |
| [features/notes.md](features/notes.md) | The microblog: notes, likes, reposts, replies, sorting |
| [features/boards.md](features/boards.md) | The textboard: boards, threads, `>>N`, sage, limits, archive |
| [features/formatting.md](features/formatting.md) | The markdown subset, quoting rules, `[aa]` art blocks |
| [features/profiles.md](features/profiles.md) | Bios, text art, the settings page |
| [features/users-and-following.md](features/users-and-following.md) | Following, the following feed, the users page |
| [features/tags.md](features/tags.md) | Creating tags, tagging notes and threads, tag pages |
| [features/search.md](features/search.md) | The search bar: scopes, shortcuts, query syntax |
| [features/dms.md](features/dms.md) | DMs and group channels: real-time IRC-style chat, join/leave, invites, message expiry |
| [features/themes-and-ui.md](features/themes-and-ui.md) | Light / eyesaver / dark, layout, search bar, sidebar |
| [design.md](design.md) | The visual system: tokens, fonts, components |
| [testing.md](testing.md) | Unit tests and how features were verified end to end |
| [showcase.md](showcase.md) | The simulated community seeded on the local instance |
| [roadmap.md](roadmap.md) | Known limits and planned work |

## In one minute

```sh
make up        # build and start app + postgres at http://localhost:8080
make logs      # follow the app log
make down      # stop
```

- Register with just a username and password.
- Post **notes** on the home page; like, repost and reply to them; sort by `recent`, `top` or
  `following`.
- **Message** people privately from their profile; `[dm]` in the nav is your inbox.
- **Follow** people from their profile or the users page. **Tag** notes and threads with tags
  created on the tags page. **Search** everything from the bar under the masthead.
- Browse **boards** (`/meta/`, `/general/`, `/tech/`, `/random/`); start threads and reply with
  `>>N` references, `sage`, and `[aa]` Shift-JIS art.
- Customize your profile with a markdown bio and up to 24×80 of text art at `/settings`.
- Switch themes with the `[theme: …]` button: light → eyesaver → dark.
