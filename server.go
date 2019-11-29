package httputils

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type GracefulShutdownServer struct {
	server          *http.Server
	wg              sync.WaitGroup
	ctx             context.Context
	cancelFunc      context.CancelFunc
	shutdownTimeout time.Duration
}

func NewGracefulShutdownServer(server *http.Server, shutdownTimeout time.Duration) *GracefulShutdownServer {
	baseContext, cancelContext := context.WithCancel(context.Background())
	server.BaseContext = func(listener net.Listener) context.Context {
		return baseContext
	}

	s := &GracefulShutdownServer{
		server:          server,
		ctx:             baseContext,
		cancelFunc:      cancelContext,
		shutdownTimeout: shutdownTimeout,
	}

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.wg.Add(1)
			next.ServeHTTP(w, r)
			s.wg.Done()
		})
	}
	server.Handler = middleware(server.Handler)

	return s
}

func (s *GracefulShutdownServer) Serve(listener net.Listener) error {
	done := make(chan error)
	go s.gracefulShutdown(done)

	serveErr := s.server.Serve(listener)

	if err := <-done; err != nil && (serveErr == http.ErrServerClosed || serveErr == nil) {
		return err
	}
	return serveErr
}

func (s *GracefulShutdownServer) ListenAndServe() error {
	done := make(chan error)
	go s.gracefulShutdown(done)

	serveErr := s.server.ListenAndServe()

	if err := <-done; err != nil && (serveErr == http.ErrServerClosed || serveErr == nil) {
		return err
	}
	return serveErr
}

func (s *GracefulShutdownServer) gracefulShutdown(done chan error) {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)
	<-stop
	close(stop)
	s.cancelFunc()

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	err := s.server.Shutdown(ctx)
	s.wg.Wait()
	done <- err
	close(done)
}
