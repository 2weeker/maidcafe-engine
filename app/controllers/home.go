package controllers

import (
	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) Home(c fiber.Ctx) error {
	return ctl.renderHome(c, fiber.StatusOK, page(c, "Home"), "", nil)
}

func (ctl *Controller) renderHome(c fiber.Ctx, status int, p templates.Page, draft string, selected []string) error {
	ctx := c.Context()
	boards, err := ctl.Store.Boards(ctx)
	if err != nil {
		return err
	}
	users, err := ctl.Store.NewestUsers(ctx, ctl.Config.Sidebar.Users)
	if err != nil {
		return err
	}
	tags, err := ctl.Store.PopularTags(ctx, ctl.Config.Sidebar.Tags)
	if err != nil {
		return err
	}
	moreBoards := len(boards) > ctl.Config.Sidebar.Boards
	boards = boards[:min(len(boards), ctl.Config.Sidebar.Boards)]
	picker, err := ctl.tagPicker(c, selected)
	if err != nil {
		return err
	}
	sort := models.ParseSort(c.Query("sort"))
	pg := pager(c, ctl.Config.Pagination.Notes)
	items, err := ctl.Store.Feed(ctx, viewerID(c), sort, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, items); err != nil {
		return err
	}
	items = models.Trim(&pg, items)
	return render(c, status, templates.Home(p, templates.HomeView{
		Boards:     boards,
		MoreBoards: moreBoards,
		Items:      items,
		Sort:       sort,
		Pager:      pg,
		Draft:      draft,
		Picker:     picker,
		Users:      users,
		Tags:       tags,
	}))
}

func (ctl *Controller) tagPicker(c fiber.Ctx, selected []string) (templates.TagPicker, error) {
	names, err := ctl.Store.TagNames(c.Context())
	if err != nil {
		return templates.TagPicker{}, err
	}
	return templates.TagPicker{Names: names, Max: ctl.Config.Tags.MaxPerPost, Selected: selected}, nil
}

func (ctl *Controller) BoardsPage(c fiber.Ctx) error {
	boards, err := ctl.Store.Boards(c.Context())
	if err != nil {
		return err
	}
	return render(c, fiber.StatusOK, templates.BoardsPage(page(c, "Boards"), boards))
}

func (ctl *Controller) Profile(c fiber.Ctx) error {
	u, err := ctl.Store.UserByUsername(c.Context(), c.Params("name"))
	if err != nil {
		return err
	}
	card, err := ctl.Store.UserCard(c.Context(), viewerID(c), u.ID)
	if err != nil {
		return err
	}
	sort := models.ParseSort(c.Query("sort"))
	if sort == models.SortFollowing {
		sort = models.SortRecent
	}
	pg := pager(c, ctl.Config.Pagination.Notes)
	items, err := ctl.Store.UserFeed(c.Context(), viewerID(c), u.ID, sort, pg.Limit(), pg.Offset())
	if err != nil {
		return err
	}
	if err := pastEnd(pg, items); err != nil {
		return err
	}
	items = models.Trim(&pg, items)
	return render(c, fiber.StatusOK, templates.Profile(page(c, "~"+u.Username), templates.ProfileView{
		User:  *u,
		Card:  *card,
		Items: items,
		Sort:  sort,
		Pager: pg,
	}))
}
