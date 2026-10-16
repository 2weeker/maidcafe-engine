package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"

	"maidcafe-engine/app/config"
	"maidcafe-engine/app/controllers"
	"maidcafe-engine/app/database"
	"maidcafe-engine/app/middleware"
	"maidcafe-engine/app/models"
	"maidcafe-engine/app/realtime"
	"maidcafe-engine/app/routes"
	"maidcafe-engine/web"
)

func purgeChat(ctx context.Context, store *models.Store, keep time.Duration) {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		if n, err := store.PurgeMessages(ctx, keep); err != nil {
			log.Printf("chat purge: %v", err)
		} else if n > 0 {
			log.Printf("chat purge: removed %d messages older than %s", n, keep)
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	store := models.NewStore(pool)

	app := fiber.New(fiber.Config{
		AppName:      "maidcafe",
		ErrorHandler: controllers.ErrorHandler,
		BodyLimit:    64 * 1024,
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Get("/static*", static.New("", static.Config{FS: web.Static(), MaxAge: 3600}))
	app.Use(middleware.LoadUser(store))
	app.Use(csrf.New(csrf.Config{
		Extractor:      extractors.FromForm("_csrf"),
		CookieSecure:   cfg.Server.CookieSecure,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	}))

	hub := realtime.NewHub()
	routes.Register(app, controllers.New(store, cfg, hub))
	if cfg.Chat.RetentionHours > 0 {
		go purgeChat(ctx, store, time.Duration(cfg.Chat.RetentionHours)*time.Hour)
	}

	go func() {
		<-ctx.Done()
		hub.Close()
		app.ShutdownWithTimeout(10 * time.Second)
	}()

	if err := app.Listen(cfg.Server.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
