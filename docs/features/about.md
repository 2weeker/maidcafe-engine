# About page

`/about` is a single page whose title and text come from `config.yaml`:

```yaml
about:
  enabled: true
  title: "about"
  body: |
    Lorem ipsum dolor sit amet, ...
```

- Every newline in the body is kept as a line break, as in notes, so write each paragraph on one
  line and separate paragraphs with a blank line.
- The body uses the same [formatting](formatting.md) as notes: bold, italic, links, lists,
  `> quotes`, `>greentext`, `[spoiler]` and so on. Raw HTML is shown as text.
- The title (1-40 characters) is the page heading, the browser tab title, and the label of the
  footer link, so `title: "the café"` gives a `[the café]` link under the footer art.
- `enabled: false` removes the footer link, and `/about` returns 404.
- The shipped `config.yaml` holds lorem ipsum placeholder text. Edit it and `make restart`;
  config is read at startup.
