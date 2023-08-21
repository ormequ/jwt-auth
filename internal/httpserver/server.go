package httpserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"jwt-auth/internal/app"
	"log"
	"net/http"
	"time"
)

type Server struct {
	http.Server
}

func New(addr string, mode string, a app.App) *Server {
	gin.SetMode(mode)

	r := gin.Default()
	s := Server{http.Server{
		Addr:    addr,
		Handler: r,
	}}
	SetRoutes(r, a)
	return &s
}

func (s *Server) Listen(ctx context.Context) error {
	errCh := make(chan error)
	defer func() {
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Shutdown(shCtx); err != nil {
			log.Printf("can't close http server listening on %s: %s", s.Addr, err.Error())
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
