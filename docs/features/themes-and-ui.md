# Themes and UI

## Themes

One button in the nav, showing the current theme's name (`[light]`), cycles through three themes:

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
┌──────────────────────────────────────────────┐▌  masthead: the one box with a hard
│              eyes over lettering             │▌  offset shadow; art and nav centered,
│            [home] [light] [name]             │▌  nav under the art
└──────────────────────────────────────────────┘▌
▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
[search]  (home page only; a centered toggle that opens to [search://][ query ][all ▾][go >>])
!! error banner (when a form fails)
┌─ main column ──────────────┐ ┌─ sidebar ──────┐
│ ==[ NOTES ]==              │ │ ==[ BOARDS ]== │
│ composer + tag pickers,    │ │ ==[ USERS ]==  │
│ [recent] [top] [following] │ │ ==[ TAGS ]==   │
└────────────────────────────┘ └────────────────┘
+-----------------+
| maidcafe v0.1.0 |
+-----------------+
```

Nav: `[home] [light]` (the theme button), then `[admin]` (mods only) `[~you] [logout]` or
`[login] [register]`. Boards,
users and tags are reached from the home sidebar (`more boards >>`, `[all users]`, `[all tags]`)
and settings from `[edit profile]` on your own profile.

The sidebar appears on the home page. On wider screens it's sticky: it stays in view, 16px from
the top, as you scroll the feed. It has no scrollbar of its own, so if it is taller than the
window its lower part comes into view only once the page is scrolled to the bottom. On phones
(< 720px) the columns stack and the sidebar moves below the feed.

## Search bar

A centered `[search]` toggle under the masthead on the home page only, a plain `<details>`
element, so no script is involved. Controllers turn it on with `Page.Search`, and only `Home`
does. It starts closed. Inside is a GET form, centered and at most 640px wide: an inverted
`search://` label, a text field (*notes, threads, ~users, #tags*), a scope dropdown and a
`go >>` button. On phones the `search://` label hides to save room. The results page has no
bar of its own, so a new search starts from home. See [search.md](search.md).

## Sidebar

- **boards:** the first `sidebar.boards` boards (default 2) and a `more boards >>` button to
  `/boards`.
- **users:** the five newest members with join date and follower count, and `[all users]`.
- **tags:** the twelve most-used tags as chips, sized by use, and `[all tags]`.

Logged out, the sidebar shows only `guests.boards` boards (default 1) and a login prompt; the
users and tags panels are members-only ([accounts.md](accounts.md#logged-out-visitors)).

## Jump links

Long pages carry a centered `v bottom` button at the top and a `^ top` button at the bottom: the home
feed, board indexes, threads and archives. On the home notes feed both are right-aligned
instead: `v bottom` at the right end of the sort tabs row, `^ top` under the last note. They're plain `#bottom` / `#top` anchor links to the
page wrapper and the footer, so they work with no JavaScript and show up in the URL.

## Footer

The ASCII box under every page holds the name and running version, `maidcafe v0.1.0`, and sizes
itself to fit. The version comes from `app/version` (`Version`), which is also printed in the
startup log. Bump it there when cutting a release.

Under the box, on one line, sits an `[about]` link when the about page is on, labelled with the
configured `about.title` ([about.md](about.md)). In textboard mode, logged-out visitors also see
`[admin]`, which opens `/login`. The link for the page you are on (`/about` or `/login`) is drawn inverted, black
with white text, and carries `aria-current="page"`.

## Pagination

Wherever lists get long: `<< newer | [ page 2 ] | older >>` (or `prev`/`next`). Disabled ends
are greyed out, and the whole bar disappears when everything fits on one page. On note lists the
top bar only appears from page 2 on.
