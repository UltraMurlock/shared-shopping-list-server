package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"shared-shopping-list-server/internal/auth"
	"shared-shopping-list-server/internal/repository"
	"shared-shopping-list-server/internal/service"
	"shared-shopping-list-server/internal/shared/database"
	appHtpp "shared-shopping-list-server/internal/transport/http"
	"shared-shopping-list-server/internal/transport/http/handler"
	"shared-shopping-list-server/internal/transport/ws"
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

	hub := ws.NewHub(logger)
	var bg sync.WaitGroup

	bg.Add(1)
	go func() {
		defer bg.Done()
		hub.Run()
	}()

	publisher := ws.NewPublisher(logger, pool, eventRepo, hub)
	pubCtx, pubCancel := context.WithCancel(context.Background())
	defer pubCancel()

	bg.Add(1)
	go func() {
		defer bg.Done()
		if err := publisher.Run(pubCtx); err != nil {
			logger.Error("publisher", "error", err)
		}
	}()

	wsHandler := handler.NewWSHandler(logger, hub, jwtManager)

	router := appHtpp.NewRouter(logger, jwtManager, authHandler, userHandler, unitHandler,
		listHandler, listItemHandler, eventHandler, wsHandler)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	srvErrCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		srvErrCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-srvErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
		}
	case sig := <-sigCh:
		logger.Info("signal received", "signal", sig.String())
	}

	// --- SHUTDOWN ---
	logger.Info("shutting down")

	pubCancel()

	hub.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown", "error", err)
	}

	bg.Wait()

	logger.Info("shutdown complete")
}
