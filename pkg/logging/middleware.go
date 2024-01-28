package logging

import (
	"net/http"

	"github.com/alvii147/flagger-api/pkg/httputils"
)

// logTraffic logs HTTP traffic, including http method, URL, protocol, and status code.
func logTraffic(w *httputils.ResponseWriter, request *http.Request) {
	l := GetLogger()

	if w.StatusCode < 400 {
		l.LogInfo(request.Method, request.URL, request.Proto, w.StatusCode, http.StatusText(w.StatusCode))
	} else if w.StatusCode < 500 {
		l.LogWarn(request.Method, request.URL, request.Proto, w.StatusCode, http.StatusText(w.StatusCode))
	} else {
		l.LogError(request.Method, request.URL, request.Proto, w.StatusCode, http.StatusText(w.StatusCode))
	}
}

// LoggerMiddleware handles logging of HTTP traffic.
func LoggerMiddleware(next httputils.HandlerFunc) httputils.HandlerFunc {
	return httputils.HandlerFunc(func(w *httputils.ResponseWriter, r *http.Request) {
		defer func() {
			logTraffic(w, r)
		}()

		next.ServeHTTP(w, r)
	})
}
