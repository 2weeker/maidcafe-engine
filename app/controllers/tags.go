package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) Tags(c fiber.Ctx) error {
	return ctl.renderTags(c, fiber.StatusOK, page(c, "Tags"), "")
}

func (ctl *Controller) CreateTag(c fiber.Ctx) error {
	name := c.FormValue("name")
	tag, err := ctl.Store.CreateTag(c.Context(), middleware.CurrentUser(c).ID, name, ctl.Config.Tags.MaxLength)
	var bad models.BadTagError
	if errors.Is(err, models.ErrTagTaken) || errors.As(err, &bad) {
		p := page(c, "Tags")
		p.Error = err.Error()
		return ctl.renderTags(c, fiber.StatusUnprocessableEntity, p, name)
	}
	if err != nil {
		return err
	}
	return c.Redirect().To(templates.TagURL(tag.Name))
}

func (ctl *Controller) renderTags(c fiber.Ctx, status int, p templates.Page, draft string) error {
	sort := models.ParseTagSort(c.Query("sort"))
	pg := pager(c, ctl.Config.Pagination.Tags)
	tags, err := ctl.Store.Tags(c.Context(), sort, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, tags); err != nil {
		return err
	}
	tags = models.Trim(&pg, tags)
	return render(c, status, templates.TagsPage(p, templates.TagsView{
		Tags:   tags,
		Sort:   sort,
		Pager:  pg,
		MaxLen: ctl.Config.Tags.MaxLength,
		Draft:  draft,
	}))
}

func (ctl *Controller) TagPage(c fiber.Ctx) error {
	tag, err := ctl.Store.TagByName(c.Context(), c.Params("name"))
	if err != nil {
		return err
	}
	show := "notes"
	if c.Query("show") == "threads" {
		show = "threads"
	}
	sort := models.ParseSort(c.Query("sort"))
	if sort == models.SortFollowing {
		sort = models.SortRecent
	}
	v := templates.TagView{Tag: *tag, Show: show, Sort: sort, PostsPerPage: ctl.Config.Pagination.Posts}
	if show == "threads" {
		pg := pager(c, ctl.Config.Pagination.Threads)
		threads, err := ctl.Store.ThreadsByTag(c.Context(), tag.ID, pg.Limit(), pg.Offset())
		if err != nil {
			return err
		}
		if err := pastEnd(pg, threads); err != nil {
			return err
		}
		v.Threads, v.Pager = models.Trim(&pg, threads), pg
	} else {
		pg := pager(c, ctl.Config.Pagination.Notes)
		notes, err := ctl.Store.NotesByTag(c.Context(), viewerID(c), tag.ID, sort, pg.Limit(), pg.Offset())
		if err != nil {
			return err
		}
		if err := pastEnd(pg, notes); err != nil {
			return err
		}
		v.Notes, v.Pager = models.Trim(&pg, notes), pg
	}
	return render(c, fiber.StatusOK, templates.TagPage(page(c, "#"+tag.Name), v))
}
