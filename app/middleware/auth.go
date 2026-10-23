package middleware

import (
	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/models"
)

const (
	SessionCookie = "mc_session"
	userKey       = "user"
)

func LoadUser(store *models.Store) fiber.Handler {
	return func(c fiber.Ctx) error {
		if token := c.Cookies(SessionCookie); token != "" {
			if u, err := store.UserBySession(c.Context(), token); err == nil {
				c.Locals(userKey, u)
			}
		}
		return c.Next()
	}
}

func RequireGuest(c fiber.Ctx) error {
	if u := CurrentUser(c); u != nil {
		return c.Redirect().To("/u/" + u.Username)
	}
	return c.Next()
}

func RequireUser(c fiber.Ctx) error {
	if CurrentUser(c) == nil {
		return c.Redirect().To("/login")
	}
	return c.Next()
}

func RequirePoster(c fiber.Ctx) error {
	if CurrentUser(c) == nil {
		return c.Redirect().To("/login")
	}
	return AllowAnon(c)
}

func AllowAnon(c fiber.Ctx) error {
	if u := CurrentUser(c); u != nil && u.Banned() {
		msg := "your account is banned"
		if u.BanReason != "" {
			msg += ": " + u.BanReason
		}
		return fiber.NewError(fiber.StatusForbidden, msg)
	}
	return c.Next()
}

func RequireMod(c fiber.Ctx) error {
	u := CurrentUser(c)
	if u == nil {
		return c.Redirect().To("/login")
	}
	if !u.IsMod() {
		return fiber.NewError(fiber.StatusForbidden, "moderators only")
	}
	return c.Next()
}

func RequireAdmin(c fiber.Ctx) error {
	u := CurrentUser(c)
	if u == nil {
		return c.Redirect().To("/login")
	}
	if !u.IsAdmin() {
		return fiber.NewError(fiber.StatusForbidden, "admins only")
	}
	return c.Next()
}

func CurrentUser(c fiber.Ctx) *models.User {
	u, _ := c.Locals(userKey).(*models.User)
	return u
}
