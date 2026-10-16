package controllers

import (
	"errors"
	"log"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"

	"maidcafe-engine/app/config"
	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/app/realtime"
	"maidcafe-engine/web/templates"
)

type Controller struct {
	Store  *models.Store
	Config config.Config
	Hub    *realtime.Hub
}

func New(store *models.Store, cfg config.Config, hub *realtime.Hub) *Controller {
	return &Controller{Store: store, Config: cfg, Hub: hub}
}

func page(c fiber.Ctx, title string) templates.Page {
	return templates.Page{
		Title:  title,
		User:   middleware.CurrentUser(c),
		CSRF:   csrf.TokenFromContext(c),
		Theme:  theme(c),
		Path:   returnPath(c),
		Query:  models.CleanQuery(c.Query("q")),
		Scope:  searchScope(c.Query("in")),
		Unread: middleware.UnreadDMs(c),
	}
}

func viewerID(c fiber.Ctx) int64 {
	if u := middleware.CurrentUser(c); u != nil {
		return u.ID
	}
	return 0
}

func pager(c fiber.Ctx, perPage int) models.Pager {
	return models.NewPager(fiber.Query[int](c, "page", 1), perPage)
}

func pastEnd[T any](pg models.Pager, items []T) error {
	if pg.Page > 1 && len(items) == 0 {
		return models.ErrNotFound
	}
	return nil
}

func render(c fiber.Ctx, status int, comp templ.Component) error {
	c.Status(status)
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return comp.Render(c.Context(), c.Response().BodyWriter())
}

func ErrorHandler(c fiber.Ctx, err error) error {
	code, msg := fiber.StatusInternalServerError, "Something broke."
	var fe *fiber.Error
	switch {
	case errors.Is(err, models.ErrNotFound):
		code, msg = fiber.StatusNotFound, "Nothing here."
	case errors.As(err, &fe):
		code, msg = fe.Code, fe.Message
	default:
		log.Printf("error: %s %s: %v", c.Method(), c.Path(), err)
	}
	return render(c, code, templates.ErrorPage(page(c, "Error"), code, msg))
}
