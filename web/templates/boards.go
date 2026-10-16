package templates

import (
	"fmt"

	"maidcafe-engine/app/config"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/markup"
)

type BoardIndex struct {
	Board        *models.Board
	Threads      []models.ThreadPreview
	Pager        models.Pager
	Rules        config.BoardRules
	PostsPerPage int
	Draft        ThreadDraft
	Picker       TagPicker
}

type ThreadView struct {
	Thread       *models.Thread
	Posts        []models.Post
	Backlinks    map[int][]int
	Pager        models.Pager
	Rules        config.BoardRules
	PostsPerPage int
	Draft        string
	Sage         bool
}

func (v ThreadView) Locked() bool {
	return v.Thread.ArchivedAt != nil || threadFull(*v.Thread, v.Rules)
}

type ArchiveView struct {
	Board   *models.Board
	Threads []models.Thread
	Pager   models.Pager
}

func threadPostURL(threadID int64, no, perPage int) string {
	return PostURL(threadID, models.PageOf(no, perPage), no)
}

func quoteURL(t models.Thread, no, perPage int) string {
	u := PageURL(threadURL(t.ID), models.PageOf(t.Posts(), perPage))
	sep := "?"
	if models.PageOf(t.Posts(), perPage) > 1 {
		sep = "&"
	}
	return fmt.Sprintf("%s%squote=%d#reply", u, sep, no)
}

func refResolver(t models.Thread, perPage int) markup.RefResolver {
	return func(no int) (string, bool) {
		if no < 1 || no > t.Posts() {
			return "", false
		}
		return threadPostURL(t.ID, no, perPage), true
	}
}

func pastBumpLimit(t models.Thread, r config.BoardRules) bool { return t.Posts() >= r.BumpLimit }
func threadFull(t models.Thread, r config.BoardRules) bool    { return t.Posts() >= r.PostLimit }
