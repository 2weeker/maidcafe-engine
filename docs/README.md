# Documentation

The project overview and quick start are in the [root README](../README.md). This folder holds
the detailed docs.

| Doc | What's in it |
| --- | --- |
| [getting-started.md](getting-started.md) | Running it, `make` targets, resetting, logs, psql, throwaway stacks |
| [configuration.md](configuration.md) | Every `config.yaml` setting and environment variable |
| [architecture.md](architecture.md) | Project layout (MVC), request flow, security model |
| [database.md](database.md) | Schema, migrations, how counts and numbering stay consistent |
| [routes.md](routes.md) | Every URL and form endpoint |
| [features/accounts.md](features/accounts.md) | Email-free registration, sessions, logout |
| [features/captcha-and-rate-limits.md](features/captcha-and-rate-limits.md) | Login/register captcha, honeypot, lockouts, request and signup rate limits, closing registration |
| [features/about.md](features/about.md) | The configurable about page |
| [features/textboard-mode.md](features/textboard-mode.md) | `mode: textboard`: anonymous boards only, staff accounts, capcodes |
| [features/tripcodes.md](features/tripcodes.md) | Names and classic 2channel / futaba tripcodes, set once on `/name`; per-board default names |
| [features/notes.md](features/notes.md) | The microblog: notes, likes, reposts, replies, tags, sorting |
| [features/boards.md](features/boards.md) | The textboard: boards, threads, `>>N`, sage, limits, archive |
| [features/formatting.md](features/formatting.md) | The markdown subset, quoting rules, greentext, spoilers, `[aa]` art blocks |
| [features/profiles.md](features/profiles.md) | Bios, text art, the settings page |
| [features/users-and-following.md](features/users-and-following.md) | Following, the following feed, the users page |
| [features/tags.md](features/tags.md) | Creating tags, tagging notes and threads, tag pages |
| [features/search.md](features/search.md) | The search bar: scopes, shortcuts, query syntax |
| [features/moderation.md](features/moderation.md) | Roles, the admin panel, bans, locks, the mod log, self-edit/delete, cooldowns |
| [features/themes-and-ui.md](features/themes-and-ui.md) | Light / eyesaver / dark, layout, search bar, sidebar |
| [design.md](design.md) | The visual system: tokens, fonts, components |
| [testing.md](testing.md) | Unit tests and how features were verified end to end |
| [showcase.md](showcase.md) | The simulated community seeded on the local instance |
| [roadmap.md](roadmap.md) | Known limits and planned work |
