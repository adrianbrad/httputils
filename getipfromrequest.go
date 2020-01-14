package httputils

import (
	"net/http"
)

func GetIPFromRequest(r *http.Request) string {
	clientIP := r.Header.Get("True-Client-IP")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Real-Ip")
	}
	if clientIP == "" {
		clientIP = r.Header.Get("X-Forwarded-For")
	}
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	return clientIP
}
