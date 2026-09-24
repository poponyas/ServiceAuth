package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/poponyas/AuthService/internal/core/secrets"
	"os"
)

func main() {
	if os.Getenv("VAULT_ADDR") != "" {
		data, err := secrets.Load(context.Background(), "service-auth")
		if err != nil {
			panic(err)
		}
		_ = os.Setenv("DATABASE_URL", data["database_url"])
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL is required")
	}
	path := "file://migrations"
	if len(os.Args) > 1 {
		path = "file://" + os.Args[1]
	}
	m, err := migrate.New(path, dsn)
	if err != nil {
		panic(err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}
	fmt.Println("migrations applied")
}
