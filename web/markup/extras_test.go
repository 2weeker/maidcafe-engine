package markup

import (
	"strings"
	"testing"
)

func TestGreentext(t *testing.T) {
	got := RenderPost(">be me\n>go to cafe\nnormal\n>with **bold** and `code`\nend", fiveposts)
	for _, w := range []string{
		`<span class="greentext">&gt;be me<br>` + "\n</span>",
		`<span class="greentext">&gt;go to cafe<br>` + "\n</span>normal<br>",
		`<span class="greentext">&gt;with <strong>bold</strong> and <code>code</code><br>` + "\n</span>end",
	} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in:\n%s", w, got)
		}
	}
	if strings.Contains(got, "<blockquote>") {
		t.Errorf("greentext became a quote:\n%s", got)
	}
}

func TestGreentextDoesNotApplyTo(t *testing.T) {
	for name, in := range map[string]string{
		"post refs":     ">>3 reply",
		"board refs":    ">>>/tech/ look",
		"literal >>":    ">>not a ref",
		"quotes":        "> quoted",
		"bare >":        ">",
		"mid line":      "text >not at start",
		"code fences":   "```\n>raw\n```",
		"aa blocks":     "[aa]\n>art\n[/aa]",
		"fake markers":  "\uE000fake marker >x",
		"inside quotes": "> >inner",
	} {
		t.Run(name, func(t *testing.T) {
			if got := RenderPost(in, fiveposts); strings.Contains(got, "greentext") {
				t.Fatalf("got %s", got)
			}
		})
	}
}

func TestGreentextKeepsIndentAndLastLine(t *testing.T) {
	got := Render("   >indented\n>last")
	if want := `<span class="greentext">&gt;indented<br>` + "\n</span>" + `<span class="greentext">&gt;last</span>`; !strings.Contains(got, want) {
		t.Fatalf("got %s", got)
	}
}

func TestSpoiler(t *testing.T) {
	for in, want := range map[string]string{
		"a [spoiler]hidden **bold**[/spoiler] b": `a <span class="spoiler" tabindex="0">hidden <strong>bold</strong></span> b`,
		"[spoiler]two\nlines[/spoiler] after":    `<span class="spoiler" tabindex="0">two<br>` + "\nlines</span> after",
		"[SPOILER]caps[/Spoiler]":                `<span class="spoiler" tabindex="0">caps</span>`,
		">green [spoiler]x[/spoiler]":            `<span class="greentext">&gt;green <span class="spoiler" tabindex="0">x</span></span>`,
		"[spoiler]*open[/spoiler] em*":           `<span class="spoiler" tabindex="0">*open</span> em*`,
	} {
		if got := Render(in); !strings.Contains(got, want) {
			t.Errorf("%q: missing %q in:\n%s", in, want, got)
		}
	}
}

func TestSpoilerStaysLiteral(t *testing.T) {
	for name, in := range map[string]string{
		"unclosed":     "[spoiler]never ends",
		"stray close":  "x[/spoiler]",
		"inline code":  "`[spoiler]a[/spoiler]`",
		"fence":        "```\n[spoiler]a[/spoiler]\n```",
		"inside links": "[a [spoiler] b](http://x) [/spoiler]",
	} {
		t.Run(name, func(t *testing.T) {
			got := Render(in)
			if strings.Contains(got, `class="spoiler"`) || !strings.Contains(got, "spoiler]") {
				t.Fatalf("got %s", got)
			}
		})
	}
}
