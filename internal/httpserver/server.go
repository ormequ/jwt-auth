package httpserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"jwt-auth/internal/app"
	"jwt-auth/internal/logger"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	http.Server
	log *slog.Logger
}

func New(log *slog.Logger, addr string, mode string, a app.App) *Server {
	gin.SetMode(mode)

	r := gin.New()
	r.Use(gin.Recovery())
	logMW := logger.Middleware(log)
	r.Use(func(c *gin.Context) {
		logMW(c.Request, c.Set, c.Next)
	})
	s := Server{
		Server: http.Server{
			Addr:    addr,
			Handler: r,
		},
		log: log,
	}
	SetRoutes(r.Group("/api"), a)
	return &s
}

func (s *Server) Listen(ctx context.Context) error {
	errCh := make(chan error)
	defer func() {
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Shutdown(shCtx); err != nil {
			s.log.Error("can't close http server listening", slog.String("addr", s.Addr), slog.String("error", err.Error()))
		}
		close(errCh)
	}()

	go func() {
		if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return fmt.Errorf("http server can't listen and serve requests: %w", err)
	}
}
