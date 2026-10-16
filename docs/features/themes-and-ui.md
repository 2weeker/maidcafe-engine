# Themes and UI

## Themes

One button in the nav, `[theme: light]`, cycles through three themes:

```
light  ->  eyesaver  ->  dark  ->  light ...
```

| Theme | Look |
| --- | --- |
| **light** (default) | Black ink on white |
| **eyesaver** | A warm coffee palette for tired eyes: espresso `#2c1b12` text on cream `#e6d8c3` boxes over a latte `#d6c4a8` page, mocha `#634533` for muted text |
| **dark** | Light grey `#ececec` on near-black `#0b0b0b`; the hard shadow turns light too |

- The choice is a cookie (`theme`, one year), so it works without JS and without an account.
- The button posts to `/theme` and returns you to the exact page you were on.
- With no cookie, the site is **light**; it doesn't follow the OS dark-mode setting.
- An old `coffee` cookie (eyesaver's earlier name) is read as eyesaver.
- The order lives in `Themes` in `web/templates/page.go`; the button follows that list.

## Page layout

```
┌──────────────────────────────────────────────┐▌  masthead: ASCII banner + nav, the one box
│  __  __   _   ___ ...      [home] [theme: …] │▌  with a hard offset shadow
└──────────────────────────────────────────────┘▌
▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
[search://][ notes, threads, ~users, #tags   ][all ▾][go >>]
!! error banner (when a form fails)
┌─ main column ──────────────┐ ┌─ sidebar ──────┐
│ ==[ NOTES ]==              │ │ ==[ BOARDS ]== │
│ composer + tag pickers,    │ │ ==[ USERS ]==  │
│ [recent] [top] [following] │ │ ==[ TAGS ]==   │
└────────────────────────────┘ └────────────────┘
+--------------------------------------+
| maidcafe :: text only :: no js :: ok |
+--------------------------------------+
```

Nav: `[home] [boards] [users] [tags] [theme: …]`, then `[dm] [~you] [settings] [logout]` (`[dm 3]`
inverted when you have unread messages) or `[login] [register]`.

The sidebar appears on the home page. On phones (< 720px) the columns stack and the sidebar
moves below the feed.

## Search bar

A full-width GET form under the masthead on every page: an inverted `search://` label, a text
field (*notes, threads, ~users, #tags*), a scope dropdown and a `go >>` button. The field keeps
your query on the results page. On phones the `search://` label hides to save room. See
[search.md](search.md).

## Sidebar

- **boards:** the first `sidebar.boards` boards (default 2) and a `more boards >>` button to
  `/boards`.
- **users:** the five newest members with join date and follower count, and `[all users]`.
- **tags:** the twelve most-used tags as chips, sized by use, and `[all tags]`.

## Pagination

Wherever lists get long: `<< newer | [ page 2 ] | older >>` (or `prev`/`next`). Disabled ends
are greyed out, and the whole bar disappears when everything fits on one page. On note lists the
top bar only appears from page 2 on.
