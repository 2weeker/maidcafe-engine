package controllers

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

const nameCookie = "mc_name"

func (ctl *Controller) identity(c fiber.Ctx) models.Identity {
	if !ctl.Config.Textboard() {
		return models.Identity{}
	}
	raw, err := base64.RawURLEncoding.DecodeString(c.Cookies(nameCookie))
	if err != nil {
		return models.Identity{}
	}
	name, key, _ := strings.Cut(string(raw), "\n")
	return models.NewIdentity(name, key)
}

func (ctl *Controller) identityView(c fiber.Ctx, anonName string) templates.IdentityView {
	id := ctl.identity(c)
	return templates.IdentityView{Name: id.Name, Trip: id.Trip, Default: anonName}
}

func (ctl *Controller) nameForm(c fiber.Ctx, status int, p templates.Page, name string) error {
	boards, err := ctl.Store.Boards(c.Context())
	if err != nil {
		return err
	}
	v := templates.NameView{Draft: name, Identity: ctl.identityView(c, ctl.Config.Boards.Defaults.AnonName)}
	if name == "" {
		v.Draft = v.Identity.Name
	}
	for _, b := range boards {
		if n := ctl.Config.Boards.For(b.Slug).AnonName; n != ctl.Config.Boards.Defaults.AnonName {
			v.BoardNames = append(v.BoardNames, [2]string{b.Slug, n})
		}
	}
	return render(c, status, templates.NamePage(p, v))
}

func (ctl *Controller) NameForm(c fiber.Ctx) error {
	return ctl.nameForm(c, fiber.StatusOK, ctl.page(c, "name & tripcode"), "")
}

func (ctl *Controller) SetName(c fiber.Ctx) error {
	name, key := c.FormValue("name"), c.FormValue("key")
	if c.FormValue("action") == "clear" {
		ctl.setNameCookie(c, "", 0)
		return c.Redirect().To("/name")
	}
	id := models.NewIdentity(name, key)
	if id.Anonymous() {
		p := ctl.page(c, "name & tripcode")
		if strings.TrimSpace(name) == "" {
			p.Error = "type a tripcode key to get a tripcode, or press clear to post anonymously."
		} else {
			p.Error = "a name needs a tripcode key. without one you post anonymously, so nobody can borrow a name."
		}
		return ctl.nameForm(c, fiber.StatusUnprocessableEntity, p, name)
	}
	if i := strings.IndexByte(name, '#'); i >= 0 && strings.TrimSpace(key) == "" {
		name, key = name[:i], strings.TrimSpace(name[i+1:])
	}
	raw := base64.RawURLEncoding.EncodeToString([]byte(models.CleanName(name) + "\n" + strings.TrimSpace(key)))
	ctl.setNameCookie(c, raw, 365)
	return c.Redirect().To("/name")
}

func (ctl *Controller) setNameCookie(c fiber.Ctx, value string, days int) {
	expires := time.Now().AddDate(0, 0, days)
	if value == "" {
		expires = time.Now().Add(-time.Hour)
	}
	c.Cookie(&fiber.Cookie{
		Name:     nameCookie,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HTTPOnly: true,
		Secure:   ctl.Config.Server.CookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}
