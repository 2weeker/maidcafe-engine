package controllers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/captcha"
	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) RegisterForm(c fiber.Ctx) error {
	if !ctl.Config.Registration.Enabled {
		return ctl.registrationClosed(c, fiber.StatusOK)
	}
	return render(c, fiber.StatusOK, templates.Register(ctl.page(c, "Register"), ctl.authForm(c, captcha.Register, "")))
}

func (ctl *Controller) registrationClosed(c fiber.Ctx, status int) error {
	return render(c, status, templates.RegistrationClosed(ctl.page(c, "Register"), ctl.Config.Registration.ClosedMessage))
}

func (ctl *Controller) registerError(c fiber.Ctx, status int, username, msg string) error {
	p := ctl.page(c, "Register")
	p.Error = msg
	return render(c, status, templates.Register(p, ctl.authForm(c, captcha.Register, username)))
}

func (ctl *Controller) Register(c fiber.Ctx) error {
	if !ctl.Config.Registration.Enabled {
		return ctl.registrationClosed(c, fiber.StatusForbidden)
	}
	username, password := c.FormValue("username"), c.FormValue("password")
	if g := ctl.gate(c, captcha.Register); g != nil {
		return ctl.registerError(c, g.status, username, g.msg)
	}
	addr := clientKey(c)
	if wait, ok := ctl.Guards.Registrations.Ready(addr); !ok {
		c.Set(fiber.HeaderRetryAfter, seconds(wait))
		return ctl.registerError(c, fiber.StatusTooManyRequests, username, waitMessage("too many new accounts from your address.", wait))
	}
	if password != c.FormValue("password_confirm") {
		return ctl.registerError(c, fiber.StatusUnprocessableEntity, username, "passwords do not match")
	}

	u, err := ctl.Store.CreateUser(c.Context(), username, password)
	if errors.Is(err, models.ErrBadUsername) || errors.Is(err, models.ErrBadPassword) || errors.Is(err, models.ErrUsernameTaken) {
		return ctl.registerError(c, fiber.StatusUnprocessableEntity, username, err.Error())
	}
	if err != nil {
		return err
	}
	ctl.Guards.Registrations.Allow(addr)
	if err := ctl.startSession(c, u.ID); err != nil {
		return err
	}
	return c.Redirect().To("/u/" + u.Username)
}

func (ctl *Controller) LoginForm(c fiber.Ctx) error {
	return render(c, fiber.StatusOK, templates.Login(ctl.page(c, "Log in"), ctl.authForm(c, captcha.Login, "")))
}

func (ctl *Controller) loginError(c fiber.Ctx, status int, username, msg string) error {
	p := ctl.page(c, "Log in")
	p.Error = msg
	return render(c, status, templates.Login(p, ctl.authForm(c, captcha.Login, username)))
}

func (ctl *Controller) Login(c fiber.Ctx) error {
	username := c.FormValue("username")
	if g := ctl.gate(c, captcha.Login); g != nil {
		return ctl.loginError(c, g.status, username, g.msg)
	}
	u, err := ctl.Store.Authenticate(c.Context(), username, c.FormValue("password"))
	if errors.Is(err, models.ErrBadLogin) {
		return ctl.loginError(c, fiber.StatusUnauthorized, username, err.Error())
	}
	if err != nil {
		return err
	}
	if err := ctl.startSession(c, u.ID); err != nil {
		return err
	}
	return c.Redirect().To("/")
}

func (ctl *Controller) About(c fiber.Ctx) error {
	if !ctl.Config.About.Enabled {
		return models.ErrNotFound
	}
	return render(c, fiber.StatusOK, templates.About(ctl.page(c, ctl.Config.About.Title), ctl.Config.About.Title, ctl.Config.About.Body))
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
