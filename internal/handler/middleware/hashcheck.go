package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/koyif/metrics/pkg/logger"
)

func WithHashCheck(hashKey string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, "/debug/pprof") {
				h.ServeHTTP(w, r)
				return
			}

			headerHash := r.Header.Get("HashSHA256")
			if headerHash == "" {
				logger.Log.Warn("hash is not provided", logger.String("URI", r.RequestURI))
				http.Error(w, "hash is not provided", http.StatusBadRequest)
				return
			}

			b, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("error reading request body", logger.Error(err))
				http.Error(w, "read error", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(b))
			r.Body.Close()

			hh := hmac.New(sha256.New, []byte(hashKey))
			hh.Write(b)
			sum := fmt.Sprintf("%x", hh.Sum(nil))

			if sum != headerHash {
				logger.Log.Warn("hash is not valid", logger.String("URI", r.RequestURI))
				http.Error(w, "hash is not valid", http.StatusBadRequest)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
