package main

import (
	"context"
	"fmt"
	"log"
	"os"
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
	"maidcafe-engine/app/routes"
	"maidcafe-engine/app/version"
	"maidcafe-engine/web"
)

func purgeAccounts(ctx context.Context, store *models.Store) {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		if names, err := store.PurgeDeletedAccounts(ctx); err != nil {
			log.Printf("account purge: %v", err)
		} else if len(names) > 0 {
			log.Printf("account purge: deleted %v", names)
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

	if len(os.Args) > 1 {
		if os.Args[1] != "adduser" {
			fmt.Fprintln(os.Stderr, "usage: maidcafe [adduser <username> [user|mod]]")
			os.Exit(2)
		}
		if err := addUser(ctx, os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("maidcafe %s, mode: %s", version.String(), cfg.Mode)

	pool, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	store := models.NewStore(pool)
	if err := store.SyncAdmins(ctx, cfg.Admin.Usernames); err != nil {
		log.Fatalf("admins: %v", err)
	}
	log.Printf("admins: %v", cfg.Admin.Usernames)

	fcfg := fiber.Config{
		AppName:      "maidcafe",
		ErrorHandler: controllers.ErrorHandler,
		BodyLimit:    64 * 1024,
	}
	if h := cfg.Server.ProxyHeader; h != "" {
		fcfg.ProxyHeader = h
		fcfg.TrustProxy = true
		fcfg.TrustProxyConfig = fiber.TrustProxyConfig{Proxies: cfg.Server.TrustedProxies}
		if len(cfg.Server.TrustedProxies) == 0 {
			fcfg.TrustProxyConfig = fiber.TrustProxyConfig{Loopback: true, Private: true}
		}
		log.Printf("proxy: client address from %s", h)
	}
	app := fiber.New(fcfg)
	app.Use(recover.New())
	app.Use(logger.New())
	app.Get("/static*", static.New("", static.Config{FS: web.Static(), MaxAge: 3600}))
	app.Use(middleware.LoadUser(store))
	rl := cfg.RateLimits
	app.Use(middleware.RateLimit(middleware.Window(rl.Guests.Requests, rl.Guests.PerSeconds), middleware.Window(rl.Users.Requests, rl.Users.PerSeconds)))
	app.Use(csrf.New(csrf.Config{
		Extractor:      extractors.FromForm("_csrf"),
		CookieSecure:   cfg.Server.CookieSecure,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	}))

	routes.Register(app, controllers.New(store, cfg))
	go purgeAccounts(ctx, store)

	go func() {
		<-ctx.Done()
		app.ShutdownWithTimeout(10 * time.Second)
	}()

	if err := app.Listen(cfg.Server.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
