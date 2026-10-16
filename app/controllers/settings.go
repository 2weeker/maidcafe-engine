package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/web/templates"
)

func (ctl *Controller) Settings(c fiber.Ctx) error {
	u := middleware.CurrentUser(c)
	return render(c, fiber.StatusOK, templates.Settings(page(c, "Settings"), *u, false))
}

func (ctl *Controller) SaveSettings(c fiber.Ctx) error {
	draft := *middleware.CurrentUser(c)
	p := page(c, "Settings")

	bio, err := models.CleanBio(c.FormValue("bio"))
	if err == nil {
		draft.Bio = bio
		draft.Art, err = models.CleanArt(c.FormValue("art"))
	}
	if err != nil {
		if !errors.Is(err, models.ErrBadBio) && !errors.Is(err, models.ErrBadArt) {
			return err
		}

		draft.Bio, draft.Art = c.FormValue("bio"), c.FormValue("art")
		p.Error = err.Error()
		return render(c, fiber.StatusUnprocessableEntity, templates.Settings(p, draft, false))
	}

	if c.FormValue("action") == "preview" {
		return render(c, fiber.StatusOK, templates.Settings(p, draft, true))
	}
	if err := ctl.Store.UpdateProfile(c.Context(), draft.ID, draft.Bio, draft.Art); err != nil {
		return err
	}
	return c.Redirect().To("/u/" + draft.Username)
}
