# Profiles

Every user has a page at `/u/:name`:

```
┌──────────────────────────────────────────┐
│  (text art, exactly as typed)            │
│ - - - - - - - - - - - - - - - - - - - -  │
│  ~username                      [follow] │
│  bio, rendered as markdown               │
│  member since 2026-10-16 06:17           │
└──────────────────────────────────────────┘
==[ NOTES ]==
[recent] [top]
notes, replies (with "re: ~x") and reposts
```

On your own profile the button is replaced by `[edit profile]`. `follow` is described in
[users-and-following.md](users-and-following.md).

## Settings (`/settings`)

Reached from `[settings]` in the nav. Two fields:

| Field | Limit | Rendering |
| --- | --- | --- |
| Bio | 500 characters | Markdown ([formatting.md](formatting.md)); may be empty |
| Text art | 24 lines × 80 columns | Exactly as typed, in monospace |

Two buttons:

- **preview** shows the profile header with your changes above the form, without saving. It's
  the first button, so pressing Enter previews rather than saves.
- **save** stores both fields and takes you to your profile.

Under the art box is a column ruler (`----+----1----+----2 … ----8`) that lines up with the
textarea so you can see the 80-column limit.

Below the profile form is the **delete account** section; see
[accounts.md](accounts.md#deleting-your-account).

## Text art rules

- **Kept:** leading spaces (they're part of the picture), blank lines inside the art.
- **Trimmed:** trailing spaces on each line, blank lines at the start and end.
- **Converted:** tabs become 4 spaces; `\r\n` becomes `\n`.
- **Removed:** control characters, including terminal color codes.
- **Refused** (with your input kept): more than 24 lines (*art is 25 lines tall (max 24)*), or a
  line over 80 columns (*art line 5 is 93 columns wide (max 80)*).

ASCII, box drawing (`┌─┐│└┘`), block shading (`░▒▓█`) and braille (`⣿`) all work. Width is
counted in characters, so double-width characters (CJK, emoji) can exceed 80 visual columns.

## Display

- The art sits at the top of the profile box in a `<pre>`, separated from the name by a dashed rule.
- It never wraps. On narrow screens the font shrinks only as much as needed: art is grouped by
  its widest line (≤ 40, ≤ 60, ≤ 80 columns) and scaled so that many columns fit, up to 14px.
- Like everything else, it's escaped text: `<script>` in art is just characters.
