package markup

import (
	"bytes"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const linkRel = "nofollow ugc noopener noreferrer"

var md = goldmark.New(
	goldmark.WithParser(parser.NewParser(
		parser.WithBlockParsers(without(parser.DefaultBlockParsers(),
			parser.NewSetextHeadingParser(),
			parser.NewATXHeadingParser(),
			parser.NewCodeBlockParser(),
			parser.NewHTMLBlockParser(),
		)...),
		parser.WithInlineParsers(append(without(parser.DefaultInlineParsers(),
			parser.NewRawHTMLParser(),
		), util.Prioritized(spoilerParser{}, 150))...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
		parser.WithASTTransformers(util.Prioritized(postProcess{}, 100)),
	)),
	goldmark.WithExtensions(extension.Strikethrough, extension.Linkify),
	goldmark.WithRendererOptions(html.WithHardWraps(), renderer.WithNodeRenderers(util.Prioritized(extraRenderer{}, 100))),
)

type RefResolver func(no int) (url string, ok bool)

var resolverKey = parser.NewContextKey()

func Render(src string) string {
	return render(normalize(src, false), nil)
}

func RenderPost(src string, resolve RefResolver) string {
	return render(normalize(src, true), resolve)
}

func render(norm string, resolve RefResolver) string {
	pc := parser.NewContext()
	if resolve != nil {
		pc.Set(resolverKey, resolve)
	}
	var buf bytes.Buffer
	if err := md.Convert([]byte(norm), &buf, parser.WithContext(pc)); err != nil {
		return "<p>" + string(util.EscapeHTML([]byte(norm))) + "</p>"
	}
	return buf.String()
}

func PostRefs(src string) []int {
	var refs []int
	for _, m := range postPlaceholderRe.FindAllStringSubmatch(normalize(src, true), -1) {
		if n, err := strconv.Atoi(m[1]); err == nil && !slices.Contains(refs, n) {
			refs = append(refs, n)
		}
	}
	slices.Sort(refs)
	return refs
}

const (
	postScheme   = "post:"
	boardScheme  = "board:"
	threadScheme = "thread:"
)

var (
	refRe             = regexp.MustCompile(`>>>/([a-z0-9]{1,16})/(\d{1,12})?|>>(\d+)`)
	postPlaceholderRe = regexp.MustCompile(`\]\(` + postScheme + `(\d+)\)`)
	aaRe              = regexp.MustCompile(`(?ms)^[ \t]*\[aa\]\n?(.*?)\n?\[/aa\]`)
)

func linkRefs(line string, posts bool) string {
	if !strings.Contains(line, ">>") {
		return line
	}
	parts := strings.Split(line, "`")
	for i := 0; i < len(parts); i += 2 {
		parts[i] = refRe.ReplaceAllStringFunc(parts[i], func(m string) string {
			sm := refRe.FindStringSubmatch(m)
			switch {
			case sm[1] != "" && sm[2] != "":
				return "[" + m + "](" + threadScheme + sm[2] + ")"
			case sm[1] != "":
				return "[" + m + "](" + boardScheme + sm[1] + ")"
			case posts && len(sm[3]) <= 6:
				return "[" + m + "](" + postScheme + sm[3] + ")"
			}
			return m
		})
	}
	return strings.Join(parts, "`")
}

func expandAA(src string) string {
	return aaRe.ReplaceAllStringFunc(src, func(m string) string {
		body := aaRe.FindStringSubmatch(m)[1]
		fence := strings.Repeat("`", max(3, longestRun(body, '`')+1))
		return "\n" + fence + "aa\n" + body + "\n" + fence + "\n"
	})
}

func longestRun(s string, c byte) int {
	best, run := 0, 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			run++
			best = max(best, run)
		} else {
			run = 0
		}
	}
	return best
}

func normalize(src string, posts bool) string {
	lines := strings.Split(expandAA(strings.ReplaceAll(src, greenMark, "")), "\n")
	out := make([]string, 0, len(lines))
	var fence string
	prevQuote := false

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)

		if fence != "" {
			if indent < 4 && strings.HasPrefix(trimmed, fence) && strings.Trim(strings.TrimSpace(trimmed), fence[:1]) == "" {
				fence = ""
			}
			out = append(out, line)
			continue
		}
		if indent < 4 {
			if f := fenceOpener(trimmed); f != "" {
				fence = f
				prevQuote = false
				out = append(out, line)
				continue
			}
		}

		isQuote := indent < 4 && (trimmed == ">" || strings.HasPrefix(trimmed, "> "))
		line = linkRefs(line, posts)
		trimmed = strings.TrimLeft(line, " ")
		if indent < 4 && strings.HasPrefix(trimmed, ">") && !isQuote {
			mark := ""
			if !strings.HasPrefix(trimmed, ">>") {
				mark = greenMark
			}
			line = strings.Repeat(" ", indent) + mark + `\` + trimmed
		}
		if prevQuote && !isQuote && strings.TrimSpace(line) != "" {
			out = append(out, "")
		}
		out = append(out, line)
		prevQuote = isQuote
	}
	return strings.Join(out, "\n")
}

func fenceOpener(s string) string {
	for _, c := range []string{"`", "~"} {
		n := len(s) - len(strings.TrimLeft(s, c))
		if n >= 3 {
			return strings.Repeat(c, n)
		}
	}
	return ""
}

type postProcess struct{}

func (postProcess) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	wrapGreentext(doc, reader.Source())
	resolve, _ := pc.Get(resolverKey).(RefResolver)
	var images []*ast.Image
	var dead []*ast.Link
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Link:
			if url, isRef, ok := resolveRef(string(n.Destination), resolve); isRef {
				if !ok {
					dead = append(dead, n)
					return ast.WalkSkipChildren, nil
				}
				n.Destination = []byte(url)
				n.SetAttributeString("class", []byte("ref"))
				return ast.WalkSkipChildren, nil
			}
			n.SetAttributeString("rel", []byte(linkRel))
		case *ast.AutoLink:
			n.SetAttributeString("rel", []byte(linkRel))
		case *ast.Image:
			images = append(images, n)
		}
		return ast.WalkContinue, nil
	})

	for _, l := range dead {
		var label []byte
		for c := l.FirstChild(); c != nil; c = c.NextSibling() {
			if t, ok := c.(*ast.Text); ok {
				label = append(label, t.Segment.Value(reader.Source())...)
			}
		}
		l.Parent().ReplaceChild(l.Parent(), l, ast.NewString(label))
	}

	for _, img := range images {
		link := ast.NewLink()
		link.Destination = img.Destination
		link.SetAttributeString("rel", []byte(linkRel))
		link.AppendChild(link, ast.NewString([]byte("[img: ")))
		for c := img.FirstChild(); c != nil; {
			next := c.NextSibling()
			link.AppendChild(link, c)
			c = next
		}
		link.AppendChild(link, ast.NewString([]byte("]")))
		img.Parent().ReplaceChild(img.Parent(), img, link)
	}
}

func resolveRef(dest string, resolve RefResolver) (url string, isRef, ok bool) {
	switch {
	case strings.HasPrefix(dest, boardScheme):
		return "/b/" + dest[len(boardScheme):], true, true
	case strings.HasPrefix(dest, threadScheme):
		return "/t/" + dest[len(threadScheme):], true, true
	case strings.HasPrefix(dest, postScheme):
		n, err := strconv.Atoi(dest[len(postScheme):])
		if err != nil || resolve == nil {
			return "", true, false
		}
		url, ok := resolve(n)
		return url, true, ok
	}
	return "", false, false
}

func without[T any](list []util.PrioritizedValue, drop ...T) []util.PrioritizedValue {
	out := list[:0]
	for _, pv := range list {
		keep := true
		for _, d := range drop {
			if reflect.TypeOf(pv.Value) == reflect.TypeOf(d) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, pv)
		}
	}
	return out
}
