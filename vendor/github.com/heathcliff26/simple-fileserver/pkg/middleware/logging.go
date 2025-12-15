package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Wrapper around ResponseWriter to ensure the status code is saved for later usage
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

// Save the written code locally after writing it to the actual ResponseWriter
func (res *responseWrapper) WriteHeader(statusCode int) {
	res.ResponseWriter.WriteHeader(statusCode)
	res.statusCode = statusCode
}

// This middleware writes information about the request to the log after it has been answered.
// Used log level: debug
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()

		wrapped := &responseWrapper{
			ResponseWriter: res,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, req)

		slog.Debug("Got Request",
			slog.String("source", ReadUserIP(req)),
			slog.Int("status", wrapped.statusCode),
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.Any("took", time.Since(start)),
		)
	})
}

// Reads the user IP from the request. Takes proxies into account.
// Reads in order from:
//
//	x-real-ip Header
//	x-forwarded-for Header
//	RemoteAddr stored in the request
func ReadUserIP(req *http.Request) string {
	IPAddress := req.Header.Get("x-real-ip")
	if IPAddress == "" {
		IPAddress = req.Header.Get("x-forwarded-for")
	}
	if IPAddress == "" {
		IPAddress = req.RemoteAddr
	}
	return IPAddress
}
