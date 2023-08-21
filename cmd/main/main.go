package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
	"jwt-auth/internal/adapters/bcrypt"
	repo "jwt-auth/internal/adapters/mongo"
	"jwt-auth/internal/app"
	"jwt-auth/internal/config"
	"jwt-auth/internal/httpserver"
	"jwt-auth/internal/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()

	eg, ctx := errgroup.WithContext(context.Background())
	log := logger.Create(cfg.Env)
	log.Info("starting app", slog.String("env", cfg.Env))

	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoConn))
	if err == nil {
		err = conn.Ping(ctx, nil)
	}
	if err != nil {
		log.Error("cannot connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	a := app.New(
		repo.New(conn.Database(cfg.MongoDB)),
		bcrypt.New(cfg.BCryptCost),
		cfg.AccessSecret,
		time.Duration(cfg.AccessExpires)*time.Second,
		time.Duration(cfg.RefreshExpires)*time.Second,
	)

	srv := httpserver.New(log, cfg.HTTPAddr, cfg.Env, a)
	sigQuit := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)

	eg.Go(func() error {
		select {
		case s := <-sigQuit:
			return fmt.Errorf("captured signal: %v", s)
		case <-ctx.Done():
			return nil
		}
	})

	eg.Go(func() (err error) {
		return srv.Listen(ctx)
	})
	if err := eg.Wait(); err != nil {
		log.Error("caught error for graceful shutdown", slog.String("error", err.Error()))
	}
	log.Info("server has been shutdown successfully")
}
