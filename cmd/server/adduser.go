package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"maidcafe-engine/app/config"
	"maidcafe-engine/app/database"
	"maidcafe-engine/app/models"
)

const addUserUsage = `usage: maidcafe adduser <username> [user|mod]

creates an account directly in the database. the password is read from the
PASSWORD environment variable, or from the first line of standard input.
accounts named under admin.usernames in config.yaml become admins at once.`

func addUser(ctx context.Context, args []string) error {
	if len(args) < 1 || len(args) > 2 || args[0] == "-h" || args[0] == "--help" {
		return errors.New(addUserUsage)
	}
	username, role := args[0], models.RoleUser
	if len(args) == 2 {
		role = args[1]
	}
	if role != models.RoleUser && role != models.RoleMod {
		return errors.New(addUserUsage)
	}
	password, err := readPassword()
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	store := models.NewStore(pool)
	u, err := store.CreateUser(ctx, username, password)
	if err != nil {
		return err
	}
	if role == models.RoleMod {
		if err := store.SetRole(ctx, u.ID, role); err != nil {
			return err
		}
	}
	if err := store.SyncAdmins(ctx, cfg.Admin.Usernames); err != nil {
		return err
	}
	u, err = store.UserByUsername(ctx, u.Username)
	if err != nil {
		return err
	}
	fmt.Printf("created ~%s (%s)\n", u.Username, u.Role)
	return nil
}

func readPassword() (string, error) {
	if p := os.Getenv("PASSWORD"); p != "" {
		return p, nil
	}
	fmt.Fprint(os.Stderr, "password: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", errors.New("no password given: set PASSWORD or type one on standard input")
	}
	return strings.TrimRight(line, "\r\n"), nil
}
