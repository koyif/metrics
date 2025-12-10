package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"
	"strings"

	"github.com/koyif/metrics/pkg/crypto"
	"github.com/koyif/metrics/pkg/logger"
)

// WithDecryption creates a middleware that decrypts encrypted request bodies.
// If the Content-Type is "application/octet-stream", it assumes the body is encrypted
// and decrypts it using the provided private key.
func WithDecryption(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	exceptions := []string{"/debug/pprof", "/swagger"}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, e := range exceptions {
				if strings.Contains(r.RequestURI, e) {
					h.ServeHTTP(w, r)
					return
				}
			}

			const contentTypeHeaderName = "Content-Type"
			contentType := r.Header.Get(contentTypeHeaderName)
			if contentType != "application/octet-stream" {
				logger.Log.Debug("request body is not encrypted", logger.String(contentTypeHeaderName, contentType))
				http.Error(w, "request body is not encrypted", http.StatusBadRequest)
				return
			}

			encryptedBody, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("error reading encrypted request body", logger.Error(err))
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}
			err = r.Body.Close()
			if err != nil {
				logger.Log.Error("error closing request body", logger.Error(err))
				return
			}

			decryptedBody, err := crypto.DecryptData(privateKey, encryptedBody)
			if err != nil {
				logger.Log.Error("error decrypting request body", logger.Error(err))
				http.Error(w, "failed to decrypt request body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))
			// Update Content-Type to application/json so handlers can process it normally
			r.Header.Set(contentTypeHeaderName, "application/json")

			h.ServeHTTP(w, r)
		})
	}
}