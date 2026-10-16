package controllers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) RegisterForm(c fiber.Ctx) error {
	return render(c, fiber.StatusOK, templates.Register(page(c, "Register"), ""))
}

func (ctl *Controller) Register(c fiber.Ctx) error {
	username, password := c.FormValue("username"), c.FormValue("password")
	if password != c.FormValue("password_confirm") {
		p := page(c, "Register")
		p.Error = "passwords do not match"
		return render(c, fiber.StatusUnprocessableEntity, templates.Register(p, username))
	}

	u, err := ctl.Store.CreateUser(c.Context(), username, password)
	if errors.Is(err, models.ErrBadUsername) || errors.Is(err, models.ErrBadPassword) || errors.Is(err, models.ErrUsernameTaken) {
		p := page(c, "Register")
		p.Error = err.Error()
		return render(c, fiber.StatusUnprocessableEntity, templates.Register(p, username))
	}
	if err != nil {
		return err
	}
	if err := ctl.startSession(c, u.ID); err != nil {
		return err
	}
	return c.Redirect().To("/u/" + u.Username)
}

func (ctl *Controller) LoginForm(c fiber.Ctx) error {
	return render(c, fiber.StatusOK, templates.Login(page(c, "Log in"), ""))
}

func (ctl *Controller) Login(c fiber.Ctx) error {
	username := c.FormValue("username")
	u, err := ctl.Store.Authenticate(c.Context(), username, c.FormValue("password"))
	if errors.Is(err, models.ErrBadLogin) {
		p := page(c, "Log in")
		p.Error = err.Error()
		return render(c, fiber.StatusUnauthorized, templates.Login(p, username))
	}
	if err != nil {
		return err
	}
	if err := ctl.startSession(c, u.ID); err != nil {
		return err
	}
	return c.Redirect().To("/")
}

func (ctl *Controller) Logout(c fiber.Ctx) error {
	if token := c.Cookies(middleware.SessionCookie); token != "" {
		if err := ctl.Store.DeleteSession(c.Context(), token); err != nil {
			return err
		}
	}
	ctl.setSessionCookie(c, "", time.Unix(0, 0))
	return c.Redirect().To("/")
}

func (ctl *Controller) startSession(c fiber.Ctx, userID int64) error {
	token, expires, err := ctl.Store.CreateSession(c.Context(), userID)
	if err != nil {
		return err
	}
	ctl.setSessionCookie(c, token, expires)
	return nil
}

func (ctl *Controller) setSessionCookie(c fiber.Ctx, value string, expires time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookie,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HTTPOnly: true,
		Secure:   ctl.Config.Server.CookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}
