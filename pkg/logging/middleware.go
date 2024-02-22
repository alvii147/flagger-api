package logging

import (
	"net/http"

	"github.com/alvii147/flagger-api/pkg/httputils"
)

// logTraffic logs HTTP traffic, including http method, URL, protocol, and status code.
func logTraffic(logger Logger, w *httputils.ResponseWriter, request *http.Request) {
	if w.StatusCode < 400 {
		logger.LogInfo(request.Method, request.URL, request.Proto, w.StatusCode, http.StatusText(w.StatusCode))
	} else if w.StatusCode < 500 {
		logger.LogWarn(request.Method, request.URL, request.Proto, w.StatusCode, http.StatusText(w.StatusCode))
	} else {
		logger.LogError(request.Method, request.URL, request.Proto, w.StatusCode, http.StatusText(w.StatusCode))
	}
}

// LoggerMiddleware handles logging of HTTP traffic.
func LoggerMiddleware(next httputils.HandlerFunc, logger Logger) httputils.HandlerFunc {
	return httputils.HandlerFunc(func(w *httputils.ResponseWriter, r *http.Request) {
		defer func() {
			logTraffic(logger, w, r)
		}()

		next.ServeHTTP(w, r)
	})
}
