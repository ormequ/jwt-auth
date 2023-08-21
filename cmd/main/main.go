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
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting app", slog.String("env", cfg.Env))

	eg, ctx := errgroup.WithContext(context.Background())

	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoConn))
	if err != nil {
		log.Error("cannot connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	a := app.New(repo.New(conn.Database(cfg.MongoDB)), bcrypt.New(cfg.BCryptCost))

	srv := httpserver.New(cfg.HTTPAddr, cfg.Env, a)
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

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case config.EnvRelease:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		)
	case config.EnvDebug:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		)
	}
	return log
}
