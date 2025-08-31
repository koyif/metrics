package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
)

func WithHashCheck(hashKey string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerHash := r.Header.Get("HashSHA256")
			if headerHash == "" {
				http.Error(w, "Hash is not valid", http.StatusBadRequest)
				return
			}

			b, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "read error", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(b))
			r.Body.Close()

			hh := hmac.New(sha256.New, []byte(hashKey))
			hh.Write(b)
			sum := fmt.Sprintf("%x", hh.Sum(nil))

			if sum != headerHash {
				http.Error(w, "Hash is not valid", http.StatusBadRequest)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
