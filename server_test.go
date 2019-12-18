package httputils

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	var wg sync.WaitGroup
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		defer wg.Done()

		fmt.Println("request handler: received request, waiting for shutdown")
		<-r.Context().Done()
		fmt.Println("request handler: received shutdown, done")
	})

	s := NewGracefulShutdownServer(nil, &http.Server{Handler: handler}, 1*time.Second)
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		wg.Add(1)
		defer wg.Done()

		fmt.Println("serve func: starting serve")
		require.EqualError(t, s.Serve(listener), http.ErrServerClosed.Error())
		fmt.Println("serve func: done")
	}()

	go func() {
		fmt.Println("send request func: sending")
		_, err = http.DefaultClient.Get("http://:" + strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
		require.NoError(t, err)
		fmt.Println("send request func: done")
	}()

	go func() {
		stop := make(chan os.Signal)
		signal.Notify(stop, syscall.SIGTERM)
		signal.Notify(stop, syscall.SIGINT)
		<-stop
		close(stop)
		fmt.Println("shutdown signal func: received signal to shutdown")
		require.NoError(t, s.Close())
		fmt.Println("shutdown signal func: done")
	}()

	time.Sleep(2 * time.Second)
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(syscall.SIGINT))
	fmt.Println("main func: sent int signal")
	wg.Wait()
	fmt.Println("main func: done")
}
