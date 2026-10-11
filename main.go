package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"shared-shopping-list-server/internal/auth"
	"shared-shopping-list-server/internal/database"
	"shared-shopping-list-server/internal/repository"
	"shared-shopping-list-server/internal/service"
	appHtpp "shared-shopping-list-server/internal/transport/http"
	"shared-shopping-list-server/internal/transport/http/handler"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

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

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("database: %v", err)
	}
	log.Println("database: migrations applied")

	jwtManager, err := auth.NewJWTManager([]byte(os.Getenv("JWT_SECRET")), time.Hour)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}

	userRepo := repository.NewUserRepo()
	userService := service.NewUserService(pool, userRepo)
	authService := service.NewAuthService(pool, userRepo, jwtManager)

	authHandler := handler.NewAuthHandler(logger, authService)
	userHandler := handler.NewUserHandler(logger, userService)

	router := appHtpp.NewRouter(logger, jwtManager, authHandler, userHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	logger.Info("server starting", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
