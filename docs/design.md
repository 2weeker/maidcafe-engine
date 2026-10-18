# Design

Black-and-white web 1.0 bones with a neo-brutalist skin: thick borders, one hard offset shadow,
ASCII ornaments, and the Inter variable font pushed to its extremes. All styling is in
`web/static/css/style.css`; there are no classes from a framework.

## Tokens

Defined on `:root` and redefined per theme (`:root[data-theme="…"]`):

| Token | Light | Eyesaver | Dark | Used for |
| --- | --- | --- | --- | --- |
| `--ink` | `#000` | `#2c1b12` | `#ececec` | Text, borders, buttons, the shadow |
| `--paper` | `#fff` | `#e6d8c3` | `#0b0b0b` | Box and input backgrounds |
| `--bg` | = paper | `#d6c4a8` | = paper | Page background |
| `--dim` | `#555` | `#634533` | `#8c8c8c` | Muted text, timestamps, quotes |
| `--line` | 3px | | | Masthead border |
| `--shadow` | `6px 6px 0 var(--ink)` | | | The masthead's hard shadow |
| `--radius` | 4px | | | Every rounded corner |
| `--mono` | system monospace stack | | | Meta lines, nav, code, art |
| `--sans` | Inter | | | Body text |

Every color comes from these tokens, so a theme is just a block that reassigns them.

## Type

- **Inter variable** (`web/static/fonts/inter.woff2` + italic), weight axis 100-900 with
  optical sizing on. Weight is the main design lever:
  - 900 with tight tracking: page titles (`.huge`)
  - 850: thread titles on the board index
  - 800: buttons, authors, board slugs
  - 420: body text
  - 200: subtitles (`/tech/ Tech`)
  - 300: search placeholder
- **Monospace** (system stack) for anything that should read as a terminal: nav, meta lines,
  section heads, hints, pagination, code, profile art.
- **AA stack** (MS PGothic and friends) only inside `[aa]` blocks: 16px with 18px line height.

## Ornaments

- ASCII banner (`MAIDCAFE` in figlet "small") in the masthead and a boxed footer, both rendered
  from Go string constants into `<pre>`.
- Section heads: `==[ NOTES ]==`.
- Bracketed links: `[home]`, `[>>]`, `[archive]`, `[view thread]`.
- Dashed rules separate a post's meta line, body and action bar.

## Components

| Component | Look |
| --- | --- |
| **Masthead** | The only box with the hard shadow: 3px border, square corners |
| **Box** (`.box`) | 2px solid ink border, 4px radius, paper background |
| **Placeholder box** | Dashed border: "not real yet" (sidebar users/tags) |
| **Buttons** | Solid ink with paper text, 2px border, 4px radius; hover inverts |
| **Secondary button** | Paper with ink text (`preview`) |
| **Action pills** (`.act`) | 1px border, mono 13px: `↳ 3` reply, `↰ 1` repost, `<3 5` like; solid ink when active. Notes and posts have no rules between header, body and actions, just spacing |
| **Tabs** | Bordered pills, the current one solid |
| **Tags** | 1px border, tiny caps: `OP`, `sage` (dashed), `full`, `bump limit`, `archived` |
| **Quotes** | 3px left rule, muted text |
| **Code** | 1px border, mono; blocks scroll sideways |
| **`>>N` links** | Mono, bold, dotted underline |
| **Alert** | Inverted ink box with `!!` prefix |
| **Search bar** | One bordered strip: inverted label, borderless input, dropdown with a CSS-drawn chevron, inverted button |

## Responsive

- Max width 920px, 16px side gutters.
- Below 720px the home columns stack (sidebar after the feed).
- The banner and profile art scale with the viewport.
- Tables drop the timestamp column under 600px.
- No horizontal page scroll: code blocks and art scroll inside their own boxes.
