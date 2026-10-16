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

// @Title      		Shared Shopping List API
// @Version         1.0
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
	unitRepo := repository.NewUnitRepo()
	listRepo := repository.NewListRepo()
	listItemRepo := repository.NewListItemRepo()
	eventRepo := repository.NewEventRepo()

	userService := service.NewUserService(pool, userRepo)
	authService := service.NewAuthService(pool, userRepo, jwtManager)
	unitService := service.NewUnitService(pool, unitRepo)
	listService := service.NewListService(pool, listRepo, eventRepo)
	listItemService := service.NewListItemService(pool, listRepo, listItemRepo, unitRepo, eventRepo)
	eventService := service.NewEventService(pool, listRepo, eventRepo)

	authHandler := handler.NewAuthHandler(logger, authService)
	userHandler := handler.NewUserHandler(logger, userService)
	unitHandler := handler.NewUnitHandler(logger, unitService)
	listHandler := handler.NewListHandler(logger, listService)
	listItemHandler := handler.NewListItemHandler(logger, listItemService)
	eventHandler := handler.NewEventHandler(logger, eventService)

	router := appHtpp.NewRouter(logger, jwtManager, authHandler, userHandler, unitHandler,
		listHandler, listItemHandler, eventHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	logger.Info("server starting", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
