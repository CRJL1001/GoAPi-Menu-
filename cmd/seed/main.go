package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	Database "github.com/m/Database"
)

//go:embed seed.sql
var seedSQL string

func seed(ctx context.Context, url string) error {
	if url == "" {
		return fmt.Errorf("TEST_DATABASE_URL est obligatoire")
	}
	db, err := Database.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	if _, err := tx.Exec(ctx, seedSQL); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := seed(ctx, os.Getenv("TEST_DATABASE_URL")); err != nil {
		log.Fatalf("Seed impossible : %v", err)
	}
	log.Println("Seed terminé : 3 ingrédients, 3 recettes et 2 menus [seed]")
}
