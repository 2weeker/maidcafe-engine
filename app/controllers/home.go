package controllers

import (
	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) Home(c fiber.Ctx) error {
	return ctl.renderHome(c, fiber.StatusOK, ctl.page(c, "Home"), "", nil)
}

func (ctl *Controller) renderHome(c fiber.Ctx, status int, p templates.Page, draft string, selected []string) error {
	ctx := c.Context()
	viewer := viewerID(c)
	guest := viewer == 0
	boards, err := ctl.Store.Boards(ctx)
	if err != nil {
		return err
	}
	boardLimit := ctl.Config.Sidebar.Boards
	var (
		users  []models.UserCard
		tags   []models.Tag
		picker templates.TagPicker
		pg     models.Pager
	)
	if guest {
		boardLimit = ctl.Config.Guests.Boards
		pg = models.NewPager(1, ctl.Config.Guests.Notes)
	} else {
		if users, err = ctl.Store.NewestUsers(ctx, ctl.Config.Sidebar.Users); err != nil {
			return err
		}
		if tags, err = ctl.Store.PopularTags(ctx, ctl.Config.Sidebar.Tags); err != nil {
			return err
		}
		if picker, err = ctl.tagPicker(c, selected); err != nil {
			return err
		}
		pg = pager(c, ctl.Config.Pagination.Notes)
	}
	moreBoards := len(boards) > boardLimit
	boards = boards[:min(len(boards), boardLimit)]
	sort := models.ParseSort(c.Query("sort"))
	items, err := ctl.Store.Feed(ctx, viewer, sort, pg.Limit(), pg.Offset())
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
	return render(c, fiber.StatusOK, templates.BoardsPage(ctl.page(c, "Boards"), boards))
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
	return render(c, fiber.StatusOK, templates.Profile(ctl.page(c, "~"+u.Username), templates.ProfileView{
		User:  *u,
		Card:  *card,
		Items: items,
		Sort:  sort,
		Pager: pg,
	}))
}
