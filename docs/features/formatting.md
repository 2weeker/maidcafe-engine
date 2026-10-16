# Formatting

User text (notes, board posts, bios) is stored exactly as typed and formatted on display by
`web/markup`, a restricted [goldmark](https://github.com/yuin/goldmark) (CommonMark) setup.
Every text box shows a one-line reminder:

```
**bold** *italic* ~~strike~~ `code` ```block``` [link](url) > quote >>1 [aa]art[/aa]
```

## Supported

| Write | Get |
| --- | --- |
| `**bold**` | **bold** |
| `*italic*` | *italic* |
| `~~strike~~` | ~~strike~~ |
| `` `code` `` | inline code |
| ```` ```lang ```` … ```` ``` ```` | a code block |
| `[text](https://…)` | a link |
| `https://…` | an automatic link |
| `- item` / `1. item` | lists |
| `> text` | a quote |
| a newline | a line break (every newline is kept) |

Leading spaces at the start of a line collapse, as in any markdown. For little doodles that need
their alignment, use a code fence or `[aa]`. A `*` is read as emphasis, so kaomoji like
`(*´∀｀*)` need `\*` to keep their asterisks.

## Deliberately not supported

| Write | Get | Why |
| --- | --- | --- |
| `# heading` | the literal text `# heading` | `#` is common on textboards; no giant text |
| `<b>html</b>` | the literal text | raw HTML is never rendered |
| 4-space indented text | normal text | only fenced code makes code blocks |
| `![alt](url)` | a link reading `[img: alt]` | the site is text-only |
| `[x](javascript:…)` | a link with an empty address | dangerous URL schemes are stripped |

## Quoting rules

These differ slightly from standard markdown, on purpose:

- **Only `> ` with a space, or a bare `>` line, makes a quote.** `>text` without a space stays
  literal, so 4chan-style greentext lines render as plain text.
- `> > text` nests a quote.
- **A plain line right after a quote starts a new paragraph.** In standard markdown,
  `> they said` followed by `my reply` would pull your reply into the quote ("lazy
  continuation"); here it doesn't.
- None of this applies inside code blocks.

## Board-only: `>>N`

In thread posts, `>>12` links to post #12 of the same thread and records a backlink (see
[boards.md](boards.md)). In notes and bios it stays literal. `>>>/board/` and `>>>/board/42`
cross-links work everywhere.

## `[aa]` Shift-JIS art

2channel's ASCII art is drawn for a *proportional* Japanese font (MS PGothic) at 16px with 18px
line spacing, so it breaks in a monospace font. Wrap it in `[aa]` … `[/aa]`:

```
[aa]
　　∧＿∧
　（　´∀｀）　＜　hello
　（　　　　）
　｜ ｜　|
　（_＿)＿）
[/aa]
```

- `[aa]` must start a line (spaces before it are fine). Mentioning `[aa]` mid-sentence or in
  `` `inline code` `` doesn't open a block. `[/aa]` can end anywhere.
- Whitespace is kept exactly.
- Quotes, `>>N` and markdown are **not** processed inside.
- It works in notes and posts.
- Font stack: MS PGothic → IPAMonaPGothic → Mona → Monapo → Textar → Saitamaar → IPAPGothic →
  Noto Sans CJK JP. Without one of the AA-metric fonts, wide pieces can drift slightly.
- An unclosed `[aa]` is just text.

Profile text art is different: it's monospace and has its own field ([profiles.md](profiles.md)).

## Safety

- HTML is escaped; raw HTML parsing is turned off.
- Links get `rel="nofollow ugc noopener noreferrer"`; internal `>>` links get `class="ref"`.
- Rendering happens on every view. Stored text is never altered except for trimming outer
  whitespace and normalizing line endings.
