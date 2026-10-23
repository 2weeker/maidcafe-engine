package templates

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/a-h/templ"

	"maidcafe-engine/app/models"
	"maidcafe-engine/app/version"
)

type Page struct {
	Title      string
	User       *models.User
	CSRF       string
	Error      string
	Theme      string
	Path       string
	Query      string
	Scope      string
	Search     bool
	GraceHours int
}

type ThreadDraft struct {
	Title string
	Body  string
	Tags  []string
	Staff bool
}

type TagPicker struct {
	Names    []string
	Max      int
	Selected []string
}

func (tp TagPicker) At(i int) string {
	if i < len(tp.Selected) {
		return tp.Selected[i]
	}
	return ""
}

func (tp TagPicker) Slots() []int {
	out := make([]int, tp.Max)
	for i := range out {
		out[i] = i
	}
	return out
}

const bannerEyes = `⠤⠤⠤⠤⠤⠤⢤⣄⣀⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠉⠙⠒⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠤⠤⠶⠶⠶⠦⠤⠤⠤⠤⠤⢤⣤⣀⣀⣀⣀⣀⣀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⢀⠄⢂⣠⣭⣭⣕⠄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠤⠀⠀⠀⠤⠀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠉⠉⠉⠉⠉⠉⠉⠉
⠀⠀⢀⠜⣳⣾⡿⠛⣿⣿⣿⣦⡠⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠠⣤⣤⣤⣤⣤⣤⣤⣤⣤⣍⣀⣦⠦⠄⣀⠀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠠⣄⣽⣿⠋⠀⡰⢿⣿⣿⣿⣿⣦⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⣿⡿⠛⠛⡿⠿⣿⣿⣿⣿⣿⣿⣷⣶⣿⣁⣂⣤⡄⠀⠀⠀⠀⠀⠀
⢳⣶⣼⣿⠃⠀⢀⠧⠤⢜⣿⣿⣿⣿⣷⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣾⠟⠁⠀⠀⠀⡇⠀⣀⡈⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⡀⠁⠐⠀⣀⠀⠀
⠀⠙⠻⣿⠀⠀⠀⠀⠀⠀⢹⣿⣿⡝⢿⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⡿⠋⠀⠀⠀⠀⠠⠃⠁⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣿⡿⠋⠀⠀
⠀⠀⠀⠙⡄⠀⠀⠀⠀⠀⢸⣿⣿⡃⢼⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⣿⣿⣿⣿⡏⠉⠉⠻⣿⡿⠋⠀⠀⠀⠀
⠀⠀⠀⠀⢰⠀⠀⠰⡒⠊⠻⠿⠋⠐⡼⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣿⣿⣿⣿⠀⠀⠀⠀⣿⠇⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠸⣇⡀⠀⠑⢄⠀⠀⠀⡠⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢖⠠⠤⠤⠔⠙⠻⠿⠋⠱⡑⢄⠀⢠⠟⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠈⠉⠒⠒⠻⠶⠛⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⡄⠀⠀⠀⠀⠀⠀⠀⠀⠡⢀⡵⠃⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⠦⣀⠀⠀⠀⠀⠀⢀⣤⡟⠉⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠉⠉⠉⠉⠙⠛⠓⠒⠲⠿⢍⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀`

const banner = ` __  __    _    ___  ___    ___    _    ___  ___
|  \/  |  /_\  |_ _||   \  / __|  /_\  | __|| __|
| |\/| | / _ \  | | | |) || (__  / _ \ | _| | _|
|_|  |_|/_/ \_\|___||___/  \___|/_/ \_\|_|  |___|`

var eyeLines = strings.Split(bannerEyes, "\n")

func eyeBaseline(i int) string {
	return strconv.FormatFloat(float64(i)*11.5+9, 'f', 1, 64)
}

func footerArt() string {
	text := "maidcafe " + version.String()
	rule := "+" + strings.Repeat("-", utf8.RuneCountInString(text)+2) + "+"
	return rule + "\n| " + text + " |\n" + rule
}

func current(p Page, path string) templ.Attributes {
	if here, _, _ := strings.Cut(p.Path, "?"); here == path {
		return templ.Attributes{"aria-current": "page"}
	}
	return nil
}

var Themes = []string{"light", "eyesaver", "dark"}

func NextTheme(t string) string {
	return Themes[(slices.Index(Themes, t)+1)%len(Themes)]
}

const artRuler = "----+----1----+----2----+----3----+----4----+----5----+----6----+----7----+----8"

func artWidthClass(art string) string {
	widest := 0
	for _, l := range strings.Split(art, "\n") {
		widest = max(widest, utf8.RuneCountInString(l))
	}
	switch {
	case widest <= 40:
		return "cols-40"
	case widest <= 60:
		return "cols-60"
	}
	return "cols-80"
}

func stamp(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04")
}

func userURL(name string) string   { return "/u/" + name }
func boardURL(slug string) string  { return "/b/" + slug }
func threadURL(id int64) string    { return fmt.Sprintf("/t/%d", id) }
func postAnchor(no int) string     { return fmt.Sprintf("p%d", no) }
func threadAnchor(id int64) string { return fmt.Sprintf("t%d", id) }
func itoa(n int) string            { return fmt.Sprint(n) }
func plural(n int, w string) string {
	if n == 1 {
		return "1 " + w
	}
	if strings.HasSuffix(w, "y") && !strings.ContainsAny(w[len(w)-2:len(w)-1], "aeiou") {
		return fmt.Sprintf("%d %sies", n, w[:len(w)-1])
	}
	return fmt.Sprintf("%d %ss", n, w)
}

func PageURL(base string, page int) string {
	if page <= 1 {
		return base
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%spage=%d", base, sep, page)
}

func NoteURL(id int64) string   { return fmt.Sprintf("/n/%d", id) }
func TagURL(name string) string { return "/tags/" + name }

func postEditURL(threadID int64, no int) string {
	return fmt.Sprintf("/t/%d/p/%d/edit", threadID, no)
}

func postDeleteURL(threadID int64, no int) string {
	return fmt.Sprintf("/t/%d/p/%d/delete", threadID, no)
}

func urlQuery(s string) string { return url.QueryEscape(s) }

func tagWeight(uses, top int) string {
	if top <= 0 {
		return "w1"
	}
	switch r := float64(uses) / float64(top); {
	case r >= 0.75:
		return "w4"
	case r >= 0.5:
		return "w3"
	case r >= 0.25:
		return "w2"
	}
	return "w1"
}

func withParam(base, key, value string) string {
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + key + "=" + value
}

func NoteAnchor(id int64, repostedBy string) string {
	if repostedBy != "" {
		return fmt.Sprintf("n%d-r-%s", id, repostedBy)
	}
	return fmt.Sprintf("n%d", id)
}

func sortURL(base string, s models.Sort) string {
	if s == models.SortRecent {
		return base
	}
	return base + "?sort=" + string(s)
}

func onOff(on bool) string {
	if on {
		return "1"
	}
	return "0"
}

func PostURL(threadID int64, page int, no int) string {
	return PageURL(threadURL(threadID), page) + "#" + postAnchor(no)
}
