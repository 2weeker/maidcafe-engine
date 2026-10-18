package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"go.yaml.in/yaml/v3"

	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) AdminHome(c fiber.Ctx) error {
	stats, err := ctl.Store.Stats(c.Context())
	if err != nil {
		return err
	}
	log, err := ctl.Store.ModLog(c.Context(), 10, 0)
	if err != nil {
		return err
	}
	return render(c, fiber.StatusOK, templates.AdminHome(ctl.page(c, "admin"), stats, log))
}

func (ctl *Controller) AdminUsers(c fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	pg := pager(c, ctl.Config.Pagination.Users)
	users, err := ctl.Store.AdminUsers(c.Context(), q, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, users); err != nil {
		return err
	}
	users = models.Trim(&pg, users)
	return render(c, fiber.StatusOK, templates.AdminUsers(ctl.page(c, "admin :: users"), templates.AdminUsersView{Users: users, Query: q, Pager: pg}))
}

func (ctl *Controller) adminBack(c fiber.Ctx) string {
	if q := c.FormValue("q"); q != "" {
		return "/admin/users?q=" + url.QueryEscape(q)
	}
	return "/admin/users"
}

func (ctl *Controller) SetRole(c fiber.Ctx) error {
	target, err := ctl.Store.UserByUsername(c.Context(), c.Params("name"))
	if err != nil {
		return err
	}
	if target.Role == models.RoleAdmin {
		return fiber.NewError(fiber.StatusForbidden, "admins are set in config.yaml")
	}
	role := c.FormValue("role")
	if err := ctl.Store.SetRole(c.Context(), target.ID, role); errors.Is(err, models.ErrBadRole) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	} else if err != nil {
		return err
	}
	ctl.Store.Log(c.Context(), middleware.CurrentUser(c).ID, "set role", "~"+target.Username, role)
	return c.Redirect().To(ctl.adminBack(c))
}

func (ctl *Controller) Ban(c fiber.Ctx) error {
	return ctl.setBan(c, true)
}

func (ctl *Controller) Unban(c fiber.Ctx) error {
	return ctl.setBan(c, false)
}

func (ctl *Controller) setBan(c fiber.Ctx, on bool) error {
	me := middleware.CurrentUser(c)
	target, err := ctl.Store.UserByUsername(c.Context(), c.Params("name"))
	if err != nil {
		return err
	}
	if target.ID == me.ID || target.Role == models.RoleAdmin || (target.Role == models.RoleMod && !me.IsAdmin()) {
		return fiber.NewError(fiber.StatusForbidden, "you can't ban that account")
	}
	reason := c.FormValue("reason")
	if err := ctl.Store.SetBan(c.Context(), target.ID, on, reason); err != nil {
		return err
	}
	action := "unban"
	if on {
		action = "ban"
	}
	ctl.Store.Log(c.Context(), me.ID, action, "~"+target.Username, reason)
	return c.Redirect().To(localPath(c.FormValue("next")))
}

func (ctl *Controller) AdminBoards(c fiber.Ctx) error {
	return ctl.renderBoards(c, fiber.StatusOK, ctl.page(c, "admin :: boards"))
}

func (ctl *Controller) renderBoards(c fiber.Ctx, status int, p templates.Page) error {
	boards, err := ctl.Store.Boards(c.Context())
	if err != nil {
		return err
	}
	return render(c, status, templates.AdminBoards(p, boards))
}

func (ctl *Controller) CreateBoard(c fiber.Ctx) error {
	slug := c.FormValue("slug")
	err := ctl.Store.CreateBoard(c.Context(), slug, c.FormValue("name"), c.FormValue("description"), formInt(c, "position"))
	if errors.Is(err, models.ErrBadBoard) || errors.Is(err, models.ErrBoardTaken) {
		p := ctl.page(c, "admin :: boards")
		p.Error = err.Error()
		return ctl.renderBoards(c, fiber.StatusUnprocessableEntity, p)
	}
	if err != nil {
		return err
	}
	ctl.Store.Log(c.Context(), middleware.CurrentUser(c).ID, "create board", "/"+strings.ToLower(strings.TrimSpace(slug))+"/", "")
	return c.Redirect().To("/admin/boards")
}

func (ctl *Controller) UpdateBoard(c fiber.Ctx) error {
	slug := c.Params("slug")
	if c.FormValue("action") == "delete" {
		err := ctl.Store.DeleteBoard(c.Context(), slug)
		if errors.Is(err, models.ErrBoardNotEmpty) {
			p := ctl.page(c, "admin :: boards")
			p.Error = err.Error()
			return ctl.renderBoards(c, fiber.StatusConflict, p)
		}
		if err != nil {
			return err
		}
		ctl.Store.Log(c.Context(), middleware.CurrentUser(c).ID, "delete board", "/"+slug+"/", "")
		return c.Redirect().To("/admin/boards")
	}
	err := ctl.Store.UpdateBoard(c.Context(), slug, c.FormValue("name"), c.FormValue("description"), formInt(c, "position"))
	if errors.Is(err, models.ErrBadBoard) {
		p := ctl.page(c, "admin :: boards")
		p.Error = err.Error()
		return ctl.renderBoards(c, fiber.StatusUnprocessableEntity, p)
	}
	if err != nil {
		return err
	}
	ctl.Store.Log(c.Context(), middleware.CurrentUser(c).ID, "edit board", "/"+slug+"/", "")
	return c.Redirect().To("/admin/boards")
}

func formInt(c fiber.Ctx, key string) int {
	var n int
	fmt.Sscan(c.FormValue(key), &n)
	return n
}

func (ctl *Controller) AdminLog(c fiber.Ctx) error {
	pg := pager(c, ctl.Config.Pagination.Users)
	log, err := ctl.Store.ModLog(c.Context(), pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, log); err != nil {
		return err
	}
	log = models.Trim(&pg, log)
	return render(c, fiber.StatusOK, templates.AdminLog(ctl.page(c, "admin :: log"), log, pg))
}

func (ctl *Controller) AdminConfig(c fiber.Ctx) error {
	cfg := ctl.Config
	if u, err := url.Parse(cfg.Database.URL); err == nil && u.User != nil {
		if _, has := u.User.Password(); has {
			u.User = url.UserPassword(u.User.Username(), "redacted")
			cfg.Database.URL = u.String()
		}
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return render(c, fiber.StatusOK, templates.AdminConfig(ctl.page(c, "admin :: config"), string(out)))
}

func (ctl *Controller) LockThread(c fiber.Ctx) error {
	return ctl.threadFlag(c, "lock", ctl.Store.SetThreadLocked)
}

func (ctl *Controller) PinThread(c fiber.Ctx) error {
	return ctl.threadFlag(c, "pin", ctl.Store.SetThreadPinned)
}

func (ctl *Controller) threadFlag(c fiber.Ctx, what string, set func(ctx context.Context, id int64, on bool) error) error {
	thread, err := ctl.Store.ThreadByID(c.Context(), fiber.Params[int64](c, "id"))
	if err != nil {
		return err
	}
	on := c.FormValue("on") == "1"
	if err := set(c.Context(), thread.ID, on); err != nil {
		return err
	}
	action := "un" + what
	if on {
		action = what
	}
	ctl.Store.Log(c.Context(), middleware.CurrentUser(c).ID, action+" thread", fmt.Sprintf("/t/%d", thread.ID), thread.Title)
	return c.Redirect().To(fmt.Sprintf("/t/%d", thread.ID))
}

func (ctl *Controller) DeleteThreadForm(c fiber.Ctx) error {
	thread, err := ctl.Store.ThreadByID(c.Context(), fiber.Params[int64](c, "id"))
	if err != nil {
		return err
	}
	return render(c, fiber.StatusOK, templates.Confirm(ctl.page(c, "delete thread"), templates.ConfirmView{
		Title:  fmt.Sprintf("delete the thread %q?", thread.Title),
		Text:   fmt.Sprintf("this removes the thread and all %d posts in it for everyone. it can't be undone.", thread.Posts()),
		Action: fmt.Sprintf("/t/%d/delete", thread.ID),
		Button: "delete thread",
		Cancel: fmt.Sprintf("/t/%d", thread.ID),
	}))
}

func (ctl *Controller) DeleteThread(c fiber.Ctx) error {
	thread, err := ctl.Store.ThreadByID(c.Context(), fiber.Params[int64](c, "id"))
	if err != nil {
		return err
	}
	if err := ctl.Store.DeleteThread(c.Context(), thread.ID); err != nil {
		return err
	}
	ctl.Store.Log(c.Context(), middleware.CurrentUser(c).ID, "delete thread", fmt.Sprintf("/b/%s", thread.BoardSlug),
		fmt.Sprintf("%q by ~%s, %d posts", thread.Title, thread.Author, thread.Posts()))
	return c.Redirect().To("/b/" + thread.BoardSlug)
}
