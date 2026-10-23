package templates

import (
	"html"
	"regexp"
	"strings"
	"testing"

	"maidcafe-engine/web/markup"
)

var tags = regexp.MustCompile(`<[^>]*>`)

func TestKaomojiRenderVerbatim(t *testing.T) {
	for _, g := range Kaomoji {
		for _, face := range g.Faces {
			got := strings.TrimSpace(html.UnescapeString(tags.ReplaceAllString(markup.RenderPost(face, nil), "")))
			if got != face {
				t.Errorf("%s: %q renders as %q", g.Name, face, got)
			}
		}
	}
}
