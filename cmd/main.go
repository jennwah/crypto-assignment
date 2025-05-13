package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jennwah/crypto-assignment/internal/config"
	"github.com/jennwah/crypto-assignment/internal/handler"
	"github.com/jennwah/crypto-assignment/internal/pkg/postgresql"
	"github.com/jennwah/crypto-assignment/internal/pkg/redis"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("failed loading application config: %w", err))
	}

	db, err := postgresql.New(cfg.Postgres)
	if err != nil {
		panic(fmt.Errorf("failed initializing connection with database: %w", err))
	}

	cache, err := redis.NewClient(cfg.Redis)
	if err != nil {
		panic(fmt.Errorf("failed initializing connection with redis Err: %v", err))
	}

	router := gin.Default()
	router.Use(gin.Recovery())

	handler.SetupHandlers(router, logger, db.DB, cache)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("listen: %w", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no params) by default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cache.Close(); err != nil {
		logger.Error("Redis close connections:", slog.Any("error", err))
	}
	if err := db.Close(); err != nil {
		logger.Error("Database close connections:", slog.Any("error", err))
	}
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server Shutdown:", slog.Any("error", err))
	}

	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	logger.Info("Graceful shutdown timeout of 5 seconds.")
	logger.Info("Server exiting")
}
