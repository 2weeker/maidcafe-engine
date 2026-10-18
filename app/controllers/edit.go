package controllers

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) loadPost(c fiber.Ctx) (*models.Thread, models.Post, error) {
	thread, err := ctl.Store.ThreadByID(c.Context(), fiber.Params[int64](c, "id"))
	if err != nil {
		return nil, models.Post{}, err
	}
	post, err := ctl.Store.PostRaw(c.Context(), thread.ID, fiber.Params[int](c, "no"))
	return thread, post, err
}

func (ctl *Controller) mayTouchPost(c fiber.Ctx, post models.Post, modOK bool) error {
	me := middleware.CurrentUser(c)
	if post.AuthorID == me.ID || (modOK && me.IsMod()) {
		return nil
	}
	return fiber.NewError(fiber.StatusForbidden, models.ErrNotYours.Error())
}

func (ctl *Controller) EditPostForm(c fiber.Ctx) error {
	thread, post, err := ctl.loadPost(c)
	if err != nil {
		return err
	}
	if err := ctl.mayTouchPost(c, post, true); err != nil {
		return err
	}
	if post.DeletedAt != nil {
		return fiber.NewError(fiber.StatusGone, models.ErrGone.Error())
	}
	return render(c, fiber.StatusOK, templates.EditPost(ctl.page(c, "edit post"), *thread, post, post.Body))
}

func (ctl *Controller) EditPost(c fiber.Ctx) error {
	thread, post, err := ctl.loadPost(c)
	if err != nil {
		return err
	}
	if err := ctl.mayTouchPost(c, post, true); err != nil {
		return err
	}
	body := c.FormValue("body")
	err = ctl.Store.EditPost(c.Context(), post.ID, body)
	switch {
	case errors.Is(err, models.ErrBadBody):
		p := ctl.page(c, "edit post")
		p.Error = err.Error()
		return render(c, fiber.StatusUnprocessableEntity, templates.EditPost(p, *thread, post, body))
	case errors.Is(err, models.ErrGone):
		return fiber.NewError(fiber.StatusGone, err.Error())
	case err != nil:
		return err
	}
	if me := middleware.CurrentUser(c); me.ID != post.AuthorID {
		ctl.Store.Log(c.Context(), me.ID, "edit post", fmt.Sprintf("/t/%d#p%d", thread.ID, post.No), "by ~"+post.Author)
	}
	return c.Redirect().To(templates.PostURL(thread.ID, models.PageOf(post.No, ctl.Config.Pagination.Posts), post.No))
}

func (ctl *Controller) DeletePostForm(c fiber.Ctx) error {
	thread, post, err := ctl.loadPost(c)
	if err != nil {
		return err
	}
	if err := ctl.mayTouchPost(c, post, true); err != nil {
		return err
	}
	back := templates.PostURL(thread.ID, models.PageOf(post.No, ctl.Config.Pagination.Posts), post.No)
	return render(c, fiber.StatusOK, templates.Confirm(ctl.page(c, "delete post"), templates.ConfirmView{
		Title:  fmt.Sprintf("delete post #%d in %q?", post.No, thread.Title),
		Text:   "the post stays in the thread as [deleted] so replies and >>" + fmt.Sprint(post.No) + " references keep working. this can't be undone by you.",
		Action: fmt.Sprintf("/t/%d/p/%d/delete", thread.ID, post.No),
		Button: "delete post",
		Cancel: back,
	}))
}

func (ctl *Controller) DeletePost(c fiber.Ctx) error {
	thread, post, err := ctl.loadPost(c)
	if err != nil {
		return err
	}
	if err := ctl.mayTouchPost(c, post, true); err != nil {
		return err
	}
	me := middleware.CurrentUser(c)
	byMod := me.ID != post.AuthorID
	if err := ctl.Store.DeletePost(c.Context(), post.ID, byMod); err != nil {
		return err
	}
	if byMod {
		ctl.Store.Log(c.Context(), me.ID, "remove post", fmt.Sprintf("/t/%d#p%d", thread.ID, post.No), "by ~"+post.Author)
	}
	return c.Redirect().To(templates.PostURL(thread.ID, models.PageOf(post.No, ctl.Config.Pagination.Posts), post.No))
}

func (ctl *Controller) RestorePost(c fiber.Ctx) error {
	thread, post, err := ctl.loadPost(c)
	if err != nil {
		return err
	}
	if err := ctl.Store.RestorePost(c.Context(), post.ID); err != nil {
		return err
	}
	ctl.Store.Log(c.Context(), middleware.CurrentUser(c).ID, "restore post", fmt.Sprintf("/t/%d#p%d", thread.ID, post.No), "by ~"+post.Author)
	return c.Redirect().To(templates.PostURL(thread.ID, models.PageOf(post.No, ctl.Config.Pagination.Posts), post.No))
}

func (ctl *Controller) loadNote(c fiber.Ctx) (int64, int64, string, error) {
	id := fiber.Params[int64](c, "id")
	owner, body, err := ctl.Store.NoteOwner(c.Context(), id)
	return id, owner, body, err
}

func (ctl *Controller) mayTouchNote(c fiber.Ctx, owner int64) error {
	me := middleware.CurrentUser(c)
	if owner == me.ID || me.IsMod() {
		return nil
	}
	return fiber.NewError(fiber.StatusForbidden, models.ErrNotYours.Error())
}

func (ctl *Controller) EditNoteForm(c fiber.Ctx) error {
	id, owner, body, err := ctl.loadNote(c)
	if err != nil {
		return err
	}
	if err := ctl.mayTouchNote(c, owner); err != nil {
		return err
	}
	return render(c, fiber.StatusOK, templates.EditNote(ctl.page(c, "edit note"), id, body))
}

func (ctl *Controller) EditNote(c fiber.Ctx) error {
	id, owner, _, err := ctl.loadNote(c)
	if err != nil {
		return err
	}
	if err := ctl.mayTouchNote(c, owner); err != nil {
		return err
	}
	body := c.FormValue("body")
	if err := ctl.Store.EditNote(c.Context(), id, body); errors.Is(err, models.ErrBadNote) {
		p := ctl.page(c, "edit note")
		p.Error = err.Error()
		return render(c, fiber.StatusUnprocessableEntity, templates.EditNote(p, id, body))
	} else if err != nil {
		return err
	}
	if me := middleware.CurrentUser(c); me.ID != owner {
		ctl.Store.Log(c.Context(), me.ID, "edit note", templates.NoteURL(id), "")
	}
	return c.Redirect().To(templates.NoteURL(id))
}

func (ctl *Controller) DeleteNoteForm(c fiber.Ctx) error {
	id, owner, _, err := ctl.loadNote(c)
	if err != nil {
		return err
	}
	if err := ctl.mayTouchNote(c, owner); err != nil {
		return err
	}
	return render(c, fiber.StatusOK, templates.Confirm(ctl.page(c, "delete note"), templates.ConfirmView{
		Title:  "delete this note?",
		Text:   "the note and every reply under it are removed for good.",
		Action: templates.NoteURL(id) + "/delete",
		Button: "delete note",
		Cancel: templates.NoteURL(id),
	}))
}

func (ctl *Controller) DeleteNote(c fiber.Ctx) error {
	id, owner, _, err := ctl.loadNote(c)
	if err != nil {
		return err
	}
	if err := ctl.mayTouchNote(c, owner); err != nil {
		return err
	}
	if err := ctl.Store.DeleteNote(c.Context(), id); err != nil {
		return err
	}
	if me := middleware.CurrentUser(c); me.ID != owner {
		ctl.Store.Log(c.Context(), me.ID, "remove note", templates.NoteURL(id), "")
	}
	return c.Redirect().To("/")
}
