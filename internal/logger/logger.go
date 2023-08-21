package logger

import (
	"context"
	"fmt"
	"jwt-auth/internal/config"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func Create(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case config.EnvDebug:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		)
	case config.EnvRelease:
		fallthrough
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		)
	}
	return log
}

func Log(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value("log").(*slog.Logger)
	if !ok {
		panic("logger has not been set")
	}
	return log
}

type MiddlewareFunc func(req *http.Request, ctxStore func(key string, value any), handle func())

func Middleware(log *slog.Logger) MiddlewareFunc {
	log.Info("logger middleware enabled")
	return func(req *http.Request, ctxStore func(key string, value any), handle func()) {
		start := time.Now()
		ctxStore("log", log)
		handle()
		log.Info(
			"request handled",
			slog.String("path", req.URL.Path),
			slog.String("query", req.URL.RawQuery),
			slog.String("method", req.Method),
			slog.String("remote_addr", req.RemoteAddr),
			slog.String("duration", fmt.Sprintf("%13v", time.Since(start))),
		)
	}
}
