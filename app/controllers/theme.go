package controllers

import (
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/web/templates"
)

const themeCookie = "theme"

func theme(c fiber.Ctx) string {
	t := c.Cookies(themeCookie)
	if t == "coffee" {
		return "eyesaver"
	}
	if slices.Contains(templates.Themes, t) {
		return t
	}
	return templates.Themes[0]
}

func (ctl *Controller) SetTheme(c fiber.Ctx) error {
	t := c.FormValue("theme")
	if !slices.Contains(templates.Themes, t) {
		return fiber.NewError(fiber.StatusBadRequest, "Unknown theme.")
	}
	c.Cookie(&fiber.Cookie{
		Name:     themeCookie,
		Value:    t,
		Path:     "/",
		Expires:  time.Now().AddDate(1, 0, 0),
		HTTPOnly: true,
		Secure:   ctl.Config.Server.CookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
	return c.Redirect().To(localPath(c.FormValue("next")))
}

func returnPath(c fiber.Ctx) string {
	if c.Method() == fiber.MethodGet {
		return c.OriginalURL()
	}
	if ref, err := url.Parse(c.Get(fiber.HeaderReferer)); err == nil && ref.Host == c.Host() {
		return ref.RequestURI()
	}
	return "/"
}

func localPath(p string) string {
	if !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") || strings.HasPrefix(p, "/\\") {
		return "/"
	}
	return p
}
