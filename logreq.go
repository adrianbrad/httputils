package httputils

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
)

func middlewareFuncLogRequest(log func (req, resp []byte)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req, _ := httputil.DumpRequest(r, true)
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