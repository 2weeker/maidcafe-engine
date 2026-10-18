package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) Users(c fiber.Ctx) error {
	sort := models.ParseUserSort(c.Query("sort"))
	pg := pager(c, ctl.Config.Pagination.Users)
	users, err := ctl.Store.Users(c.Context(), viewerID(c), sort, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, users); err != nil {
		return err
	}
	users = models.Trim(&pg, users)
	return render(c, fiber.StatusOK, templates.UsersPage(ctl.page(c, "Users"), templates.UsersView{Users: users, Sort: sort, Pager: pg}))
}

func (ctl *Controller) Followers(c fiber.Ctx) error {
	return ctl.followList(c, "followers")
}

func (ctl *Controller) FollowingList(c fiber.Ctx) error {
	return ctl.followList(c, "following")
}

func (ctl *Controller) followList(c fiber.Ctx, kind string) error {
	u, err := ctl.Store.UserByUsername(c.Context(), c.Params("name"))
	if err != nil {
		return err
	}
	pg := pager(c, ctl.Config.Pagination.Users)
	var users []models.UserCard
	if kind == "followers" {
		users, err = ctl.Store.Followers(c.Context(), viewerID(c), u.ID, pg.Limit(), pg.Offset())
	} else {
		users, err = ctl.Store.Following(c.Context(), viewerID(c), u.ID, pg.Limit(), pg.Offset())
	}
	if err != nil {
		return err
	}
	if err := pastEnd(pg, users); err != nil {
		return err
	}
	users = models.Trim(&pg, users)
	return render(c, fiber.StatusOK, templates.FollowListPage(ctl.page(c, "~"+u.Username+" "+kind),
		templates.FollowListView{User: *u, Kind: kind, Users: users, Pager: pg}))
}

func (ctl *Controller) Follow(c fiber.Ctx) error {
	target, err := ctl.Store.UserByUsername(c.Context(), c.Params("name"))
	if err != nil {
		return err
	}
	err = ctl.Store.SetFollow(c.Context(), middleware.CurrentUser(c).ID, target.ID, c.FormValue("on") == "1")
	if errors.Is(err, models.ErrSelfFollow) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err != nil {
		return err
	}
	return c.Redirect().To(localPath(c.FormValue("next")))
}
