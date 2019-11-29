package httputils

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
)

func MiddlewareFuncLogRequest(log func(req, resp []byte), response bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req, _ := httputil.DumpRequest(r, true)

			if !response {
				log(req, nil)
				next.ServeHTTP(w, r)
				return
			}

			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)

			for k, v := range rec.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(rec.Code)
			rec.Body.WriteTo(w)

			resp, _ := httputil.DumpResponse(rec.Result(), true)
			log(req, resp)
		})
	}
}
