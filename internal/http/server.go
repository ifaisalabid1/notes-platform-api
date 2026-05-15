package http

import (
	"context"
	"errors"
	"log/slog"
	"net"
	stdhttp "net/http"
	"time"
)

type Server struct {
	server *stdhttp.Server
	logger *slog.Logger
}

func NewServer(port string, handler stdhttp.Handler, logger *slog.Logger) *Server {
	return &Server{
		logger: logger,
		server: &stdhttp.Server{
			Addr:              net.JoinHostPort("", port),
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	s.logger.Info("starting http server", slog.String("addr", s.server.Addr))

	if err := s.server.ListenAndServe(); !errors.Is(err, stdhttp.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	return s.server.Shutdown(ctx)
}
