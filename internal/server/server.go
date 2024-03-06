package server

import (
	"context"
	"net/http"

	"github.com/stsg/gophermart2/internal/config"
	"github.com/stsg/gophermart2/internal/router"
	"github.com/stsg/gophermart2/internal/services/shutdowner"
	"go.uber.org/zap"
)

type Server struct {
	http http.Server
}

func New(ctx context.Context) *Server {
	return &Server{
		http: http.Server{
			Addr:    config.Get().RunAddress,
			Handler: router.New(ctx),
		},
	}
}

func Run(ctx context.Context) {
	srv := New(ctx)
	go srv.ListenAndServer()
	srv.addToShutdowner()
}

func (s *Server) ListenAndServer() {
	if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zap.L().Fatal(err.Error())
	}
}

func (s *Server) addToShutdowner() {
	shutdowner.Get().AddCloser(func(ctx context.Context) error {
		if err := s.http.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	})
}
