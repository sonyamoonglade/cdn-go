package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	s *http.Server
}

func NewServer(port string, host string, h http.Handler) *Server {
	return &Server{
		s: &http.Server{
			Addr:              fmt.Sprintf("%s:%s", host, port),
			Handler:           h,
			ReadTimeout:       time.Second * 15,
			ReadHeaderTimeout: time.Second * 15,
			WriteTimeout:      time.Second * 15,
			IdleTimeout:       time.Second * 15,
			MaxHeaderBytes:    1 << 20,
		},
	}
}

func (s *Server) ListenAndServe() error {
	err := s.s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}
