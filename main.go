package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"shared-shopping-list-server/internal/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf(".env file: %v", err)
	}

	ctx := context.Background()

	dbCfg, err := database.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("db config: %v", err)
	}

	pool, err := database.NewPool(ctx, dbCfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	defer pool.Close()
}
