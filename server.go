package httputils

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

type GracefulShutdownServer struct {
	server            *http.Server
	wg                sync.WaitGroup
	ctx               context.Context
	cancelAllRequests context.CancelFunc
	shutdownTimeout   time.Duration
}

func NewGracefulShutdownServer(baseContext context.Context, server *http.Server, shutdownTimeout time.Duration) *GracefulShutdownServer {
	s := &GracefulShutdownServer{
		server:          server,
		shutdownTimeout: shutdownTimeout,
	}

	if baseContext == nil {
		baseContext = context.Background()
	}

	s.ctx, s.cancelAllRequests = context.WithCancel(baseContext)
	s.server.BaseContext = func(listener net.Listener) context.Context {
		return s.ctx
	}

	return s
}

func (s *GracefulShutdownServer) Serve(listener net.Listener) error {
	return s.server.Serve(listener)
}

func (s *GracefulShutdownServer) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *GracefulShutdownServer) Close() error {
	s.cancelAllRequests()

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}
