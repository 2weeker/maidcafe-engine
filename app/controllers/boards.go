package controllers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/markup"
	"maidcafe-engine/web/templates"
)

func cooldownError(err error) error {
	var cd models.CooldownError
	if errors.As(err, &cd) {
		return fiber.NewError(fiber.StatusTooManyRequests, cd.Error())
	}
	return err
}

func (ctl *Controller) Board(c fiber.Ctx) error {
	board, err := ctl.Store.BoardBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return err
	}
	return ctl.renderBoard(c, fiber.StatusOK, ctl.page(c, "/"+board.Slug+"/"), board, templates.ThreadDraft{})
}

func (ctl *Controller) CreateThread(c fiber.Ctx) error {
	board, err := ctl.Store.BoardBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return err
	}
	rules := ctl.Config.Boards.For(board.Slug)
	draft := templates.ThreadDraft{Title: c.FormValue("title"), Body: c.FormValue("body"), Tags: tagFields(c, ctl.Config.Tags.MaxPerPost)}
	me := middleware.CurrentUser(c)
	if err := cooldownError(ctl.Store.ThreadQuota(c.Context(), me.ID, ctl.Config.Cooldowns.ThreadsPerDay)); err != nil {
		return err
	}
	if err := cooldownError(ctl.Store.PostCooldown(c.Context(), me.ID, time.Duration(ctl.Config.Cooldowns.PostSeconds)*time.Second)); err != nil {
		return err
	}
	tagIDs, err := ctl.Store.ResolveTags(c.Context(), draft.Tags, ctl.Config.Tags.MaxPerPost)
	var id int64
	if err == nil {
		id, err = ctl.Store.CreateThread(c.Context(), board.ID, middleware.CurrentUser(c).ID, draft.Title, draft.Body, rules.MaxThreads, tagIDs)
	}
	if errors.Is(err, models.ErrBadTitle) || errors.Is(err, models.ErrBadBody) || errors.Is(err, models.ErrUnknownTag) || errors.Is(err, models.ErrTooManyTag) {
		p := ctl.page(c, "/"+board.Slug+"/")
		p.Error = err.Error()
		return ctl.renderBoard(c, fiber.StatusUnprocessableEntity, p, board, draft)
	}
	if err != nil {
		return err
	}
	return c.Redirect().To(fmt.Sprintf("/t/%d", id))
}

func (ctl *Controller) renderBoard(c fiber.Ctx, status int, p templates.Page, board *models.Board, draft templates.ThreadDraft) error {
	rules := ctl.Config.Boards.For(board.Slug)
	pg := pager(c, ctl.Config.Pagination.Threads)
	threads, err := ctl.Store.LiveThreads(c.Context(), board.ID, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, threads); err != nil {
		return err
	}
	threads = models.Trim(&pg, threads)
	previews, err := ctl.Store.Previews(c.Context(), threads, rules.PreviewReplies)
	if err != nil {
		return err
	}
	picker, err := ctl.tagPicker(c, draft.Tags)
	if err != nil {
		return err
	}
	return render(c, status, templates.BoardPage(p, templates.BoardIndex{
		Board:        board,
		Threads:      previews,
		Pager:        pg,
		Rules:        rules,
		PostsPerPage: ctl.Config.Pagination.Posts,
		Draft:        draft,
		Picker:       picker,
	}))
}

func (ctl *Controller) Archive(c fiber.Ctx) error {
	board, err := ctl.Store.BoardBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return err
	}
	pg := pager(c, ctl.Config.Pagination.Archive)
	threads, err := ctl.Store.ArchivedThreads(c.Context(), board.ID, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, threads); err != nil {
		return err
	}
	threads = models.Trim(&pg, threads)
	return render(c, fiber.StatusOK, templates.ArchivePage(ctl.page(c, "/"+board.Slug+"/ archive"),
		templates.ArchiveView{Board: board, Threads: threads, Pager: pg}))
}

func (ctl *Controller) Thread(c fiber.Ctx) error {
	return ctl.renderThread(c, fiber.StatusOK, "", "", false)
}

func (ctl *Controller) Reply(c fiber.Ctx) error {
	id := fiber.Params[int64](c, "id")
	body, sage := c.FormValue("body"), c.FormValue("sage") == "1"
	thread, err := ctl.Store.ThreadByID(c.Context(), id)
	if err != nil {
		return err
	}
	rules := ctl.Config.Boards.For(thread.BoardSlug)
	if err := cooldownError(ctl.Store.PostCooldown(c.Context(), middleware.CurrentUser(c).ID, time.Duration(ctl.Config.Cooldowns.PostSeconds)*time.Second)); err != nil {
		return err
	}
	no, err := ctl.Store.CreateReply(c.Context(), id, middleware.CurrentUser(c).ID, body, models.ReplyOptions{
		Sage:      sage,
		BumpLimit: rules.BumpLimit,
		PostLimit: rules.PostLimit,
		Refs:      markup.PostRefs(body),
	})
	switch {
	case errors.Is(err, models.ErrBadBody):
		return ctl.renderThread(c, fiber.StatusUnprocessableEntity, err.Error(), body, sage)
	case errors.Is(err, models.ErrThreadFull), errors.Is(err, models.ErrThreadArchived), errors.Is(err, models.ErrThreadLocked):
		return ctl.renderThread(c, fiber.StatusConflict, err.Error(), body, sage)
	case err != nil:
		return err
	}
	per := ctl.Config.Pagination.Posts
	return c.Redirect().To(templates.PostURL(id, models.PageOf(no, per), no))
}

func (ctl *Controller) renderThread(c fiber.Ctx, status int, errMsg, draft string, sage bool) error {
	thread, err := ctl.Store.ThreadByID(c.Context(), fiber.Params[int64](c, "id"))
	if err != nil {
		return err
	}
	per := ctl.Config.Pagination.Posts
	pg := pager(c, per)
	posts, err := ctl.Store.PostsByThread(c.Context(), thread.ID, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, posts); err != nil {
		return err
	}
	posts = models.Trim(&pg, posts)

	backlinks := map[int][]int{}
	if len(posts) > 0 {
		if backlinks, err = ctl.Store.Backlinks(c.Context(), thread.ID, posts[0].No, posts[len(posts)-1].No); err != nil {
			return err
		}
	}
	if q := fiber.Query[int](c, "quote"); draft == "" && q >= 1 && q <= thread.Posts() {
		draft = fmt.Sprintf(">>%d\n", q)
	}

	p := ctl.page(c, thread.Title)
	p.Error = errMsg
	return render(c, status, templates.ThreadPage(p, templates.ThreadView{
		Thread:       thread,
		Posts:        posts,
		Backlinks:    backlinks,
		Pager:        pg,
		Rules:        ctl.Config.Boards.For(thread.BoardSlug),
		PostsPerPage: per,
		Draft:        draft,
		Sage:         sage,
	}))
}
