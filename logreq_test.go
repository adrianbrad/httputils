package httputils

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestMiddlewareFuncLogRequestWebsocket(t *testing.T) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool {
			return true
		},
	}
	handler := MiddlewareFuncLogRequest(
		func(req, resp []byte) {
			t.Logf("request: %s, response: %s", string(req), string(resp))
		}, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
	}))
	s := httptest.NewServer(handler)

	u, err := url.Parse(s.URL)
	require.NoError(t, err)
	u.Scheme = "ws"
	fmt.Println(u.String())
	_, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.NoError(t, err)
	assert.Equal(t, 101, resp.StatusCode)
}
