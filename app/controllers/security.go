package controllers

import (
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/captcha"
	"maidcafe-engine/app/config"
	"maidcafe-engine/app/limits"
	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

const honeypotField = templates.HoneypotField

type Guards struct {
	Captcha       *captcha.Store
	Fails         map[captcha.Purpose]*limits.Failures
	AuthAttempts  *limits.Rate
	Registrations *limits.Rate
	AnonPosts     *limits.Rate
	AnonThreads   *limits.Rate
	Honeypot      bool
}

func newGuards(cfg config.Config) Guards {
	lock := cfg.Captcha.Lockout
	lockFor := time.Duration(lock.Seconds) * time.Second
	g := Guards{
		Fails: map[captcha.Purpose]*limits.Failures{
			captcha.Login:    limits.NewFailures(lock.Failures, lockFor),
			captcha.Register: limits.NewFailures(lock.Failures, lockFor),
			captcha.Post:     limits.NewFailures(lock.Failures, lockFor),
		},
		AuthAttempts:  rate(cfg.RateLimits.AuthAttempts),
		Registrations: rate(cfg.RateLimits.Registrations),
		Honeypot:      cfg.Captcha.Honeypot,
	}
	if cfg.Textboard() {
		g.AnonPosts = limits.NewRate(1, time.Duration(cfg.Cooldowns.PostSeconds)*time.Second)
		g.AnonThreads = limits.NewRate(cfg.Cooldowns.ThreadsPerDay, 24*time.Hour)
	}
	if cfg.Captcha.Enabled {
		g.Captcha = captcha.New(captcha.DefaultCapacity, cfg.Captcha.Length, time.Duration(cfg.Captcha.LifetimeSeconds)*time.Second)
	}
	return g
}

func rate(l config.Limit) *limits.Rate {
	return limits.NewRate(l.Requests, time.Duration(l.PerSeconds)*time.Second)
}

func clientKey(c fiber.Ctx) string {
	return limits.AddressKey(c.IP())
}

func anonCooldown(c fiber.Ctx, r *limits.Rate, what string) error {
	if wait, ok := r.Ready(clientKey(c)); !ok {
		c.Set(fiber.HeaderRetryAfter, seconds(wait))
		return fiber.NewError(fiber.StatusTooManyRequests, models.CooldownError{Wait: wait, What: what}.Error())
	}
	return nil
}

func seconds(d time.Duration) string {
	return strconv.Itoa(int(math.Ceil(d.Seconds())))
}

func waitMessage(what string, d time.Duration) string {
	n := int(math.Ceil(d.Seconds()))
	switch {
	case n >= 3600:
		return what + " try again in about " + plural(int(math.Ceil(float64(n)/3600)), "hour") + "."
	case n >= 60:
		return what + " try again in about " + plural(int(math.Ceil(float64(n)/60)), "minute") + "."
	}
	return what + " try again in " + plural(n, "second") + "."
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}

func (ctl *Controller) authForm(c fiber.Ctx, purpose captcha.Purpose, username string) templates.AuthForm {
	f := templates.AuthForm{Username: username, Honeypot: ctl.Guards.Honeypot}
	if ctl.Guards.Captcha == nil {
		return f
	}
	addr := clientKey(c)
	if wait, locked := ctl.Guards.Fails[purpose].Locked(addr); locked {
		f.Locked = waitMessage("too many wrong captchas.", wait)
		return f
	}
	f.Captcha = &templates.CaptchaView{
		ID:     ctl.Guards.Captcha.Issue(captcha.Binding{Purpose: purpose, Addr: addr}),
		Width:  captcha.Width,
		Height: captcha.Height,
	}
	return f
}

type gateResult struct {
	status int
	msg    string
}

func (ctl *Controller) gate(c fiber.Ctx, purpose captcha.Purpose) *gateResult {
	if wait, ok := ctl.Guards.AuthAttempts.Allow(clientKey(c)); !ok {
		c.Set(fiber.HeaderRetryAfter, seconds(wait))
		return &gateResult{fiber.StatusTooManyRequests, waitMessage("too many attempts from your address.", wait)}
	}
	if g := ctl.trap(c, purpose); g != nil {
		return g
	}
	return ctl.checkCaptcha(c, purpose)
}

func (ctl *Controller) postGuard(c fiber.Ctx) templates.AuthForm {
	if middleware.CurrentUser(c) != nil {
		return templates.AuthForm{Honeypot: ctl.Guards.Honeypot}
	}
	return ctl.authForm(c, captcha.Post, "")
}

func (ctl *Controller) postGate(c fiber.Ctx) *gateResult {
	if g := ctl.trap(c, captcha.Post); g != nil {
		return g
	}
	if middleware.CurrentUser(c) != nil {
		return nil
	}
	return ctl.checkCaptcha(c, captcha.Post)
}

func (ctl *Controller) trap(c fiber.Ctx, purpose captcha.Purpose) *gateResult {
	if ctl.Guards.Honeypot && c.FormValue(honeypotField) != "" {
		log.Printf("honeypot: %s refused for %s", purpose, clientKey(c))
		return &gateResult{fiber.StatusBadRequest, "that didn't go through. please try again."}
	}
	return nil
}

func (ctl *Controller) checkCaptcha(c fiber.Ctx, purpose captcha.Purpose) *gateResult {
	if ctl.Guards.Captcha == nil {
		return nil
	}
	addr := clientKey(c)
	fails := ctl.Guards.Fails[purpose]
	if wait, locked := fails.Locked(addr); locked {
		c.Set(fiber.HeaderRetryAfter, seconds(wait))
		return &gateResult{fiber.StatusTooManyRequests, waitMessage("too many wrong captchas.", wait)}
	}
	b := captcha.Binding{Purpose: purpose, Addr: addr}
	if ctl.Guards.Captcha.Verify(c.FormValue("captcha_id"), c.FormValue("captcha_value"), b) {
		fails.Succeed(addr)
		return nil
	}
	if lock := fails.Fail(addr); lock > 0 {
		log.Printf("captcha: %s locked out of %s for %s", addr, purpose, lock)
		c.Set(fiber.HeaderRetryAfter, seconds(lock))
		return &gateResult{fiber.StatusTooManyRequests, waitMessage("too many wrong captchas.", lock)}
	}
	return &gateResult{fiber.StatusUnprocessableEntity, "wrong or expired captcha. here's a new one."}
}

func (ctl *Controller) CaptchaImage(c fiber.Ctx) error {
	id, ok := strings.CutSuffix(c.Params("file"), ".png")
	if !ok || ctl.Guards.Captcha == nil {
		return fiber.ErrNotFound
	}
	png, ok := ctl.Guards.Captcha.Image(id, clientKey(c))
	if !ok {
		return fiber.ErrNotFound
	}
	c.Set(fiber.HeaderCacheControl, "private, max-age=300")
	c.Set(fiber.HeaderContentType, "image/png")
	return c.Send(png)
}
