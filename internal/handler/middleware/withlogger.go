package middleware

import (
	"fmt"
	"github.com/koyif/metrics/internal/app/logger"
	"net/http"
	"time"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.responseData.size += size

	return size, err
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.ResponseWriter.WriteHeader(statusCode)
	lrw.responseData.status = statusCode
}

func WithLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rd := &responseData{status: 0, size: 0}
		responseWriter := loggingResponseWriter{
			w,
			rd,
		}

		h.ServeHTTP(&responseWriter, r)

		duration := time.Since(start)
		uri := r.RequestURI
		method := r.Method

		logger.Log.Info(
			"request log",
			logger.String("URI", uri),
			logger.String("method", method),
			logger.String("duration", fmt.Sprintf("%dµs", duration.Milliseconds())),
		)
		logger.Log.Info(
			"response log",
			logger.Int("status", rd.status),
			logger.Int("size", rd.size),
		)
	})
}
