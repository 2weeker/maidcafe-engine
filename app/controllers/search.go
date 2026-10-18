package controllers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func searchScope(s string, loggedIn bool) string {
	for _, v := range templates.SearchScopes(loggedIn) {
		if v == s {
			return v
		}
	}
	return "all"
}

func (ctl *Controller) Search(c fiber.Ctx) error {
	ctx := c.Context()
	q := models.CleanQuery(c.Query("q"))
	viewer := viewerID(c)
	guest := viewer == 0
	scope := searchScope(c.Query("in"), !guest)
	p := ctl.page(c, "Search")

	switch {
	case strings.HasPrefix(q, "#"):
		if tag, err := ctl.Store.TagByName(ctx, q); err == nil {
			return c.Redirect().To(templates.TagURL(tag.Name))
		} else if !errors.Is(err, models.ErrNotFound) {
			return err
		}
		q, scope = models.NormalizeTag(q), "tags"
	case strings.HasPrefix(q, "~") || strings.HasPrefix(q, "@"):
		if u, err := ctl.Store.UserByUsername(ctx, q[1:]); err == nil {
			return c.Redirect().To("/u/" + u.Username)
		} else if !errors.Is(err, models.ErrNotFound) {
			return err
		}
		q, scope = models.NormalizeUsername(q[1:]), "users"
	}
	if guest && (scope == "users" || scope == "tags") {
		scope = "all"
	}
	p.Query, p.Scope = q, scope
	v := templates.SearchView{Query: q, Scope: scope, PostsPerPage: ctl.Config.Pagination.Posts}
	if q == "" {
		return render(c, fiber.StatusOK, templates.SearchPage(p, v))
	}

	if scope == "all" {
		const each = 5
		var err error
		if v.Notes, err = ctl.Store.SearchNotes(ctx, viewer, q, each, 0); err != nil {
			return err
		}
		if v.Posts, err = ctl.Store.SearchPosts(ctx, q, each, 0); err != nil {
			return err
		}
		if v.Threads, err = ctl.Store.SearchThreads(ctx, q, each, 0); err != nil {
			return err
		}
		if !guest {
			if v.Users, err = ctl.Store.SearchUsers(ctx, viewer, q, each, 0); err != nil {
				return err
			}
			if v.Tags, err = ctl.Store.SearchTags(ctx, q, each, 0); err != nil {
				return err
			}
		}
		v.Each = each
		return render(c, fiber.StatusOK, templates.SearchPage(p, v))
	}

	pg := pager(c, ctl.Config.Pagination.Search)
	var err error
	switch scope {
	case "notes":
		var items []models.FeedItem
		if items, err = ctl.Store.SearchNotes(ctx, viewer, q, pg.Limit(), pg.Offset()); err == nil {
			v.Notes = models.Trim(&pg, items)
		}
	case "threads":
		var posts []models.SearchPost
		if posts, err = ctl.Store.SearchPosts(ctx, q, pg.Limit(), pg.Offset()); err == nil {
			v.Posts = models.Trim(&pg, posts)
		}
		if err == nil && pg.Page == 1 {
			v.Threads, err = ctl.Store.SearchThreads(ctx, q, 20, 0)
		}
	case "users":
		var users []models.UserCard
		if users, err = ctl.Store.SearchUsers(ctx, viewer, q, pg.Limit(), pg.Offset()); err == nil {
			v.Users = models.Trim(&pg, users)
		}
	case "tags":
		var tags []models.Tag
		if tags, err = ctl.Store.SearchTags(ctx, q, pg.Limit(), pg.Offset()); err == nil {
			v.Tags = models.Trim(&pg, tags)
		}
	}
	if err != nil {
		return err
	}
	if pg.Page > 1 && v.Empty() {
		return models.ErrNotFound
	}
	v.Pager = pg
	return render(c, fiber.StatusOK, templates.SearchPage(p, v))
}
