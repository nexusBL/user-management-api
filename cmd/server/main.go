package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/nexusBL/user-management-api/config"
	"github.com/nexusBL/user-management-api/internal/handler"
	applogger "github.com/nexusBL/user-management-api/internal/logger"
	"github.com/nexusBL/user-management-api/internal/middleware"
	"github.com/nexusBL/user-management-api/internal/repository"
	"github.com/nexusBL/user-management-api/internal/routes"
	"github.com/nexusBL/user-management-api/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, err := applogger.New(cfg.AppEnv)
	if err != nil {
		log.Fatalf("create logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	db, err := openDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("connect database", zap.Error(err))
	}
	defer db.Close()

	validate, err := handler.NewValidator()
	if err != nil {
		logger.Fatal("create validator", zap.Error(err))
	}

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(
		userRepository,
		cfg.DefaultPageLimit,
		cfg.MaxPageLimit,
		time.Now,
	)
	userHandler := handler.NewUserHandler(userService, validate)

	app := fiber.New(fiber.Config{
		AppName:      "user-management-api",
		ErrorHandler: handler.NewErrorHandler(logger),
	})

	app.Use(recovermw.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestDuration(logger))

	routes.Register(app, userHandler)

	// Keep unmatched routes consistent with the centralized JSON error format.
	app.Use(func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})

	logger.Info("starting server",
		zap.String("environment", cfg.AppEnv),
		zap.String("port", cfg.ServerPort),
	)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- app.Listen(":" + cfg.ServerPort)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-serverErrors:
		if err != nil {
			logger.Fatal("server stopped unexpectedly", zap.Error(err))
		}
	case sig := <-signals:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			logger.Fatal("graceful shutdown failed", zap.Error(err))
		}

		logger.Info("server shut down cleanly")
	}
}

func openDB(databaseURL string) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
