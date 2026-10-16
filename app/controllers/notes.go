package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) CreateNote(c fiber.Ctx) error {
	body := c.FormValue("body")
	selected := tagFields(c, ctl.Config.Tags.MaxPerPost)
	tagIDs, err := ctl.Store.ResolveTags(c.Context(), selected, ctl.Config.Tags.MaxPerPost)
	if err == nil {
		_, err = ctl.Store.CreateNote(c.Context(), middleware.CurrentUser(c).ID, body, nil, tagIDs)
	}
	if errors.Is(err, models.ErrBadNote) || errors.Is(err, models.ErrUnknownTag) || errors.Is(err, models.ErrTooManyTag) {
		p := page(c, "Home")
		p.Error = err.Error()
		return ctl.renderHome(c, fiber.StatusUnprocessableEntity, p, body, selected)
	}
	if err != nil {
		return err
	}
	return c.Redirect().To("/")
}

func (ctl *Controller) Note(c fiber.Ctx) error {
	return ctl.renderNote(c, fiber.StatusOK, "", "")
}

func (ctl *Controller) ReplyNote(c fiber.Ctx) error {
	parentID := fiber.Params[int64](c, "id")
	body := c.FormValue("body")
	id, err := ctl.Store.CreateNote(c.Context(), middleware.CurrentUser(c).ID, body, &parentID, nil)
	if errors.Is(err, models.ErrBadNote) {
		return ctl.renderNote(c, fiber.StatusUnprocessableEntity, err.Error(), body)
	}
	if err != nil {
		return err
	}
	parent, err := ctl.Store.NoteByID(c.Context(), 0, parentID)
	if err != nil {
		return err
	}
	last := models.PageOf(parent.ReplyCount, ctl.Config.Pagination.Replies)
	return c.Redirect().To(templates.PageURL(templates.NoteURL(parentID), last) + "#" + templates.NoteAnchor(id, ""))
}

func (ctl *Controller) renderNote(c fiber.Ctx, status int, errMsg, draft string) error {
	viewer := viewerID(c)
	note, err := ctl.Store.NoteByID(c.Context(), viewer, fiber.Params[int64](c, "id"))
	if err != nil {
		return err
	}
	var parent *models.Note
	if note.ReplyToID != nil {
		if parent, err = ctl.Store.NoteByID(c.Context(), viewer, *note.ReplyToID); err != nil {
			return err
		}
	}
	pg := pager(c, ctl.Config.Pagination.Replies)
	replies, err := ctl.Store.Replies(c.Context(), viewer, note.ID, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, replies); err != nil {
		return err
	}
	replies = models.Trim(&pg, replies)
	p := page(c, "~"+note.Author)
	p.Error = errMsg
	return render(c, status, templates.NotePage(p, note, parent, replies, pg, draft))
}

func tagFields(c fiber.Ctx, max int) []string {
	out := make([]string, 0, max)
	for i := 1; i <= max; i++ {
		out = append(out, c.FormValue(fmt.Sprintf("tag%d", i)))
	}
	return out
}

func (ctl *Controller) LikeNote(c fiber.Ctx) error {
	return ctl.setNoteFlag(c, ctl.Store.SetLike)
}

func (ctl *Controller) RepostNote(c fiber.Ctx) error {
	return ctl.setNoteFlag(c, ctl.Store.SetRepost)
}

func (ctl *Controller) setNoteFlag(c fiber.Ctx, set func(ctx context.Context, userID, noteID int64, on bool) error) error {
	on := c.FormValue("on") == "1"
	if err := set(c.Context(), middleware.CurrentUser(c).ID, fiber.Params[int64](c, "id"), on); err != nil {
		return err
	}
	return c.Redirect().To(localPath(c.FormValue("next")))
}
