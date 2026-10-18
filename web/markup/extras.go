package markup

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const greenMark = ""

var (
	kindGreentext   = ast.NewNodeKind("Greentext")
	kindSpoiler     = ast.NewNodeKind("Spoiler")
	kindSpoilerOpen = ast.NewNodeKind("SpoilerOpen")

	spoilerOpenTag  = []byte("[spoiler]")
	spoilerCloseTag = []byte("[/spoiler]")
)

type greentext struct{ ast.BaseInline }

func (n *greentext) Kind() ast.NodeKind { return kindGreentext }
func (n *greentext) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

type spoiler struct{ ast.BaseInline }

func (n *spoiler) Kind() ast.NodeKind { return kindSpoiler }
func (n *spoiler) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

type spoilerOpen struct {
	ast.BaseInline
	Segment text.Segment
}

func (n *spoilerOpen) Kind() ast.NodeKind { return kindSpoilerOpen }
func (n *spoilerOpen) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

type spoilerParser struct{}

func (spoilerParser) Trigger() []byte { return []byte{'['} }

func (spoilerParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, seg := block.PeekLine()
	switch {
	case hasFoldPrefix(line, spoilerOpenTag):
		block.Advance(len(spoilerOpenTag))
		return &spoilerOpen{Segment: seg.WithStop(seg.Start + len(spoilerOpenTag))}
	case hasFoldPrefix(line, spoilerCloseTag):
		var open *spoilerOpen
		for c := parent.LastChild(); c != nil; c = c.PreviousSibling() {
			if o, ok := c.(*spoilerOpen); ok {
				open = o
				break
			}
		}
		if open == nil {
			return nil
		}
		block.Advance(len(spoilerCloseTag))
		parser.ProcessDelimiters(open, pc)
		sp := &spoiler{}
		for c := open.NextSibling(); c != nil; {
			next := c.NextSibling()
			sp.AppendChild(sp, c)
			c = next
		}
		parent.RemoveChild(parent, open)
		return sp
	}
	return nil
}

func (spoilerParser) CloseBlock(parent ast.Node, block text.Reader, pc parser.Context) {
	var stray []*spoilerOpen
	ast.Walk(parent, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if o, ok := n.(*spoilerOpen); ok && entering {
			stray = append(stray, o)
		}
		return ast.WalkContinue, nil
	})
	for _, o := range stray {
		o.Parent().ReplaceChild(o.Parent(), o, ast.NewTextSegment(o.Segment))
	}
}

func hasFoldPrefix(s, prefix []byte) bool {
	return len(s) >= len(prefix) && bytes.EqualFold(s[:len(prefix)], prefix)
}

func wrapGreentext(doc *ast.Document, source []byte) {
	var starts []*ast.Text
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if t, ok := n.(*ast.Text); ok && entering && bytes.HasPrefix(t.Segment.Value(source), []byte(greenMark)) {
			starts = append(starts, t)
		}
		return ast.WalkContinue, nil
	})
	for _, t := range starts {
		t.Segment = t.Segment.WithStart(t.Segment.Start + len(greenMark))
		parent := t.Parent()
		g := &greentext{}
		parent.InsertBefore(parent, t, g)
		for c := ast.Node(t); c != nil; {
			next := c.NextSibling()
			g.AppendChild(g, c)
			if lt, ok := c.(*ast.Text); ok && (lt.SoftLineBreak() || lt.HardLineBreak()) {
				break
			}
			c = next
		}
	}
}

type extraRenderer struct{}

func (extraRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindGreentext, spanRenderer(`<span class="greentext">`))
	reg.Register(kindSpoiler, spanRenderer(`<span class="spoiler" tabindex="0">`))
}

func spanRenderer(open string) renderer.NodeRendererFunc {
	return func(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			_, _ = w.WriteString(open)
		} else {
			_, _ = w.WriteString("</span>")
		}
		return ast.WalkContinue, nil
	}
}
