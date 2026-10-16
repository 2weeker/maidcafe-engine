package markup

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func fiveposts(no int) (string, bool) {
	if no < 1 || no > 5 {
		return "", false
	}
	return fmt.Sprintf("/t/9#p%d", no), true
}

func TestRenderPostRefs(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []string
		notWant []string
	}{
		{
			name: "ref at line start links to the post",
			in:   ">>3 agreed",
			want: []string{`<a href="/t/9#p3" class="ref">&gt;&gt;3</a> agreed`},
		},
		{
			name: "ref mid line",
			in:   "see >>1 and >>2",
			want: []string{`<a href="/t/9#p1" class="ref">&gt;&gt;1</a>`, `<a href="/t/9#p2" class="ref">&gt;&gt;2</a>`},
		},
		{
			name:    "ref to a missing post is plain text",
			in:      ">>99 who?",
			want:    []string{"&gt;&gt;99 who?"},
			notWant: []string{"<a", "post:"},
		},
		{
			name: "ref inside a quote",
			in:   "> >>2 said this",
			want: []string{"<blockquote>", `<a href="/t/9#p2" class="ref">`},
		},
		{
			name:    "ref inside inline code is untouched",
			in:      "type `>>2` to link",
			want:    []string{"<code>&gt;&gt;2</code>"},
			notWant: []string{"/t/9#p2"},
		},
		{
			name:    "ref inside a code fence is untouched",
			in:      "```\n>>2\n```",
			want:    []string{"<pre><code>&gt;&gt;2\n</code></pre>"},
			notWant: []string{"<a"},
		},
		{
			name: "cross-board link to a board",
			in:   "try >>>/tech/ instead",
			want: []string{`<a href="/b/tech" class="ref">&gt;&gt;&gt;/tech/</a>`},
		},
		{
			name: "cross-board link to a thread",
			in:   ">>>/tech/42",
			want: []string{`<a href="/t/42" class="ref">&gt;&gt;&gt;/tech/42</a>`},
		},
		{
			name:    "silly long numbers are left alone",
			in:      ">>12345678",
			notWant: []string{"<a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderPost(tt.in, fiveposts)
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

func TestRenderWithoutResolverKeepsPostRefsLiteral(t *testing.T) {
	got := Render(">>3 hi")
	if strings.Contains(got, "<a") || !strings.Contains(got, "&gt;&gt;3 hi") {
		t.Fatalf("got %s", got)
	}

	if got := Render("see >>>/meta/"); !strings.Contains(got, `href="/b/meta"`) {
		t.Fatalf("got %s", got)
	}
}

func TestPostRefs(t *testing.T) {
	got := PostRefs(">>1 >>3\n> >>1 again\n`>>4`\n```\n>>5\n```\n>>>/tech/7 >>2")
	if want := []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAA(t *testing.T) {
	art := "　　∧＿∧\n　（　´∀｀）\n  >>2 > not a quote"
	got := RenderPost("look:\n[aa]\n"+art+"\n[/aa]\nnice", fiveposts)
	for _, w := range []string{
		`<pre><code class="language-aa">　　∧＿∧`,
		"  &gt;&gt;2 &gt; not a quote\n</code></pre>",
		"<p>nice</p>",
	} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in:\n%s", w, got)
		}
	}
	if strings.Contains(got, "<blockquote>") || strings.Contains(got, `class="ref"`) {
		t.Errorf("aa content was processed:\n%s", got)
	}
}

func TestAAContainingBackticks(t *testing.T) {
	got := Render("[aa]```\n``` ```[/aa]")
	if !strings.Contains(got, "<code class=\"language-aa\">```\n``` ```\n</code>") {
		t.Fatalf("got %s", got)
	}
}

func TestUnclosedAAIsLiteral(t *testing.T) {
	if got := Render("[aa] no end"); !strings.Contains(got, "[aa] no end") {
		t.Fatalf("got %s", got)
	}
}

func TestAAOnlyOpensAtLineStart(t *testing.T) {
	for name, in := range map[string]string{
		"mid sentence":   "use the [aa] tags so spacing survives\n[aa]\n x\n[/aa]",
		"in inline code": "- wrap it in `[aa]...[/aa]` for art",
	} {
		t.Run(name, func(t *testing.T) {
			got := Render(in)
			if strings.Contains(got, "use the\n") || strings.Contains(got, "<li>`</li>") {
				t.Fatalf("inline [aa] opened a block:\n%s", got)
			}
			if name == "mid sentence" && !strings.Contains(got, `<code class="language-aa"> x`) {
				t.Fatalf("line-start [aa] should still open a block:\n%s", got)
			}
			if name == "in inline code" && !strings.Contains(got, "<code>[aa]...[/aa]</code>") {
				t.Fatalf("inline code should be untouched:\n%s", got)
			}
		})
	}
}
