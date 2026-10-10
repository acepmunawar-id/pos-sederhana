// Command migrate manages the database schema.
//
//	go run ./cmd/migrate up        apply all migrations
//	go run ./cmd/migrate down [n]  roll back n migrations (default: all)
//	go run ./cmd/migrate seed      insert demo data (idempotent)
//	go run ./cmd/migrate fresh     drop everything, migrate and seed
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/posumkm/backend/internal/repository/postgres"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: migrate up|down [n]|seed|fresh")
	}
	ctx := context.Background()

	seed := func() error {
		pool, err := postgres.Connect(ctx, url, 2)
		if err != nil {
			return err
		}
		defer pool.Close()
		return postgres.Seed(ctx, pool)
	}

	switch os.Args[1] {
	case "up":
		if err := postgres.MigrateUp(url); err != nil {
			return err
		}
		fmt.Println("migrations applied")
	case "down":
		steps := 0
		if len(os.Args) > 2 {
			n, err := strconv.Atoi(os.Args[2])
			if err != nil {
				return fmt.Errorf("invalid step count %q", os.Args[2])
			}
			steps = n
		}
		if err := postgres.MigrateDown(url, steps); err != nil {
			return err
		}
		fmt.Println("migrations rolled back")
	case "seed":
		if err := seed(); err != nil {
			return err
		}
		fmt.Println("seed data inserted")
	case "fresh":
		if err := postgres.MigrateDown(url, 0); err != nil {
			return err
		}
		if err := postgres.MigrateUp(url); err != nil {
			return err
		}
		if err := seed(); err != nil {
			return err
		}
		fmt.Println("database recreated and seeded")
	default:
		return fmt.Errorf("unknown command %q (use up, down, seed or fresh)", os.Args[1])
	}
	return nil
}
