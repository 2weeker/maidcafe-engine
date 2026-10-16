package middleware

import (
	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/models"
)

const (
	SessionCookie = "mc_session"
	userKey       = "user"
	unreadKey     = "unread_dms"
)

func LoadUser(store *models.Store) fiber.Handler {
	return func(c fiber.Ctx) error {
		if token := c.Cookies(SessionCookie); token != "" {
			if u, err := store.UserBySession(c.Context(), token); err == nil {
				c.Locals(userKey, u)
				if n, err := store.UnreadMessages(c.Context(), u.ID); err == nil {
					c.Locals(unreadKey, n)
				}
			}
		}
		return c.Next()
	}
}

func RequireUser(c fiber.Ctx) error {
	if CurrentUser(c) == nil {
		return c.Redirect().To("/login")
	}
	return c.Next()
}

func UnreadDMs(c fiber.Ctx) int {
	n, _ := c.Locals(unreadKey).(int)
	return n
}

func CurrentUser(c fiber.Ctx) *models.User {
	u, _ := c.Locals(userKey).(*models.User)
	return u
}
