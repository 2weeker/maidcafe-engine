package middleware

import (
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/limits"
)

func RateLimit(guests, users *limits.Rate) fiber.Handler {
	return func(c fiber.Ctx) error {
		r, key := guests, "a:"+limits.AddressKey(c.IP())
		if u := CurrentUser(c); u != nil {
			r, key = users, "u:"+strconv.FormatInt(u.ID, 10)
		}
		wait, ok := r.Allow(key)
		if ok {
			return c.Next()
		}
		n := int(math.Ceil(wait.Seconds()))
		c.Set(fiber.HeaderRetryAfter, strconv.Itoa(n))
		return fiber.NewError(fiber.StatusTooManyRequests, "slow down: too many requests. try again in "+strconv.Itoa(n)+" "+unit(n)+".")
	}
}

func unit(n int) string {
	if n == 1 {
		return "second"
	}
	return "seconds"
}

func Window(requests, perSeconds int) *limits.Rate {
	return limits.NewRate(requests, time.Duration(perSeconds)*time.Second)
}
