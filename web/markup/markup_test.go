package markup

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []string
		notWant []string
	}{
		{
			name: "inline formatting",
			in:   "**bold** *it* `code` ~~gone~~",
			want: []string{"<strong>bold</strong>", "<em>it</em>", "<code>code</code>", "<del>gone</del>"},
		},
		{
			name:    "raw html is shown as text",
			in:      `<b>hi</b> <script>alert(1)</script>`,
			want:    []string{"&lt;b&gt;hi&lt;/b&gt;", "&lt;script&gt;"},
			notWant: []string{"<b>", "<script>", "raw HTML omitted"},
		},
		{
			name: "html block is shown as text",
			in:   "<div>\nx\n</div>",
			want: []string{"&lt;div&gt;"},
		},
		{
			name: "quote needs a space",
			in:   "> quoted",
			want: []string{"<blockquote>", "quoted"},
		},
		{
			name:    "no space is not a quote",
			in:      ">not quoted",
			want:    []string{"&gt;not quoted"},
			notWant: []string{"<blockquote>"},
		},
		{
			name: "bare > continues a quote",
			in:   "> one\n>\n> two",
			want: []string{"<blockquote>", "one", "two"},
		},
		{
			name: "line after a quote is not swallowed into it",
			in:   "> they said\nmy reply",
			want: []string{"</blockquote>\n<p>my reply</p>"},
		},
		{
			name: "nested quote",
			in:   "> > deep",
			want: []string{"<blockquote>\n<blockquote>"},
		},
		{
			name:    "quote syntax inside code fence is left alone",
			in:      "```\n>raw\n> raw2\nafter\n```",
			want:    []string{"<pre><code>&gt;raw\n&gt; raw2\nafter\n</code></pre>"},
			notWant: []string{"\\&gt;", "<blockquote>"},
		},
		{
			name: "newlines become line breaks",
			in:   "line one\nline two",
			want: []string{"line one<br>\nline two"},
		},
		{
			name:    "headings are not rendered",
			in:      "# not a heading\ntext\n---",
			notWant: []string{"<h1>", "<h2>"},
		},
		{
			name:    "indented text is not a code block",
			in:      "    indented",
			notWant: []string{"<pre>"},
		},
		{
			name:    "javascript links are neutered",
			in:      "[x](javascript:alert(1))",
			notWant: []string{"javascript:"},
		},
		{
			name: "links get rel",
			in:   "[site](https://example.com) and https://example.org",
			want: []string{
				`<a href="https://example.com" rel="nofollow ugc noopener noreferrer">site</a>`,
				`<a href="https://example.org" rel="nofollow ugc noopener noreferrer">https://example.org</a>`,
			},
		},
		{
			name:    "images become links",
			in:      "![a cat](https://example.com/cat.png)",
			want:    []string{`<a href="https://example.com/cat.png"`, "[img: a cat]"},
			notWant: []string{"<img"},
		},
		{
			name: "lists",
			in:   "- a\n- b",
			want: []string{"<ul>", "<li>a</li>"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.in)
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("missing %q in:\n%s", w, got)
				}
			}
			for _, w := range tt.notWant {
				if strings.Contains(got, w) {
					t.Errorf("unexpected %q in:\n%s", w, got)
				}
			}
		})
	}
}
