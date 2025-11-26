package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkHashComputation measures the performance of HMAC-SHA256 computation
func BenchmarkHashComputation(b *testing.B) {
	hashKey := "test-secret-key"
	payloads := []struct {
		name string
		size int
	}{
		{"small_100B", 100},
		{"medium_1KB", 1024},
		{"large_10KB", 10240},
		{"xlarge_100KB", 102400},
	}

	for _, p := range payloads {
		b.Run(p.name, func(b *testing.B) {
			data := bytes.Repeat([]byte("x"), p.size)
			b.ResetTimer()
			b.SetBytes(int64(p.size))

			for i := 0; i < b.N; i++ {
				hh := hmac.New(sha256.New, []byte(hashKey))
				hh.Write(data)
				_ = fmt.Sprintf("%x", hh.Sum(nil))
			}
		})
	}
}

// BenchmarkHashCheckMiddleware measures the full middleware performance
func BenchmarkHashCheckMiddleware(b *testing.B) {
	hashKey := "test-secret-key"
	payloads := []struct {
		name string
		size int
	}{
		{"small_100B", 100},
		{"medium_1KB", 1024},
		{"large_10KB", 10240},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, p := range payloads {
		b.Run(p.name, func(b *testing.B) {
			payload := bytes.Repeat([]byte("x"), p.size)

			hh := hmac.New(sha256.New, []byte(hashKey))
			hh.Write(payload)
			hash := fmt.Sprintf("%x", hh.Sum(nil))

			middleware := WithHashCheck(hashKey)
			wrappedHandler := middleware(handler)

			b.ResetTimer()
			b.SetBytes(int64(p.size))

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/update", io.NopCloser(bytes.NewReader(payload)))
				req.Header.Set("HashSHA256", hash)
				w := httptest.NewRecorder()

				wrappedHandler.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkHashCheckMiddlewareParallel measures concurrent middleware performance
func BenchmarkHashCheckMiddlewareParallel(b *testing.B) {
	hashKey := "test-secret-key"
	payload := []byte(`{"id":"test","type":"counter","delta":1}`)

	hh := hmac.New(sha256.New, []byte(hashKey))
	hh.Write(payload)
	hash := fmt.Sprintf("%x", hh.Sum(nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := WithHashCheck(hashKey)
	wrappedHandler := middleware(handler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/update", io.NopCloser(bytes.NewReader(payload)))
			req.Header.Set("HashSHA256", hash)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)
		}
	})
}

// BenchmarkInvalidHashCheck measures performance when hash validation fails
func BenchmarkInvalidHashCheck(b *testing.B) {
	hashKey := "test-secret-key"
	payload := []byte(`{"id":"test","type":"counter","delta":1}`)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := WithHashCheck(hashKey)
	wrappedHandler := middleware(handler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update", io.NopCloser(bytes.NewReader(payload)))
		req.Header.Set("HashSHA256", "invalid_hash")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)
	}
}

// BenchmarkBodyReadAll measures io.ReadAll performance with different payload sizes
func BenchmarkBodyReadAll(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"100B", 100},
		{"1KB", 1024},
		{"10KB", 10240},
		{"100KB", 102400},
	}

	for _, s := range sizes {
		b.Run(s.name, func(b *testing.B) {
			data := bytes.Repeat([]byte("x"), s.size)
			b.ResetTimer()
			b.SetBytes(int64(s.size))

			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(data)
				_, _ = io.ReadAll(reader)
			}
		})
	}
}
