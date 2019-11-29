package httputils

import (
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"os"
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
		<-r.Context().Done()
	})
	s := NewGracefulShutdownServer(&http.Server{Handler: handler}, 1*time.Second)
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		wg.Add(1)
		defer wg.Done()
		require.EqualError(t, s.Serve(listener), http.ErrServerClosed.Error())
	}()

	go func() {
		_, err = http.DefaultClient.Get("http://:" + strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
		require.NoError(t, err)
	}()

	time.Sleep(2 * time.Second)
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(syscall.SIGINT))

	wg.Wait()
}
