package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithIPCheck_EmptySubnet(t *testing.T) {
	middleware, err := WithIPCheck("")
	if err != nil {
		t.Fatalf("WithIPCheck(\"\") returned error: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestWithIPCheck_InvalidCIDR(t *testing.T) {
	_, err := WithIPCheck("invalid-cidr")
	if err == nil {
		t.Error("WithIPCheck with invalid CIDR should return error")
	}
}

func TestWithIPCheck_MissingHeader(t *testing.T) {
	middleware, err := WithIPCheck("192.168.1.0/24")
	if err != nil {
		t.Fatalf("WithIPCheck returned error: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestWithIPCheck_EmptyHeader(t *testing.T) {
	middleware, err := WithIPCheck("192.168.1.0/24")
	if err != nil {
		t.Fatalf("WithIPCheck returned error: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestWithIPCheck_InvalidIPFormat(t *testing.T) {
	middleware, err := WithIPCheck("192.168.1.0/24")
	if err != nil {
		t.Fatalf("WithIPCheck returned error: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "not-an-ip")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestWithIPCheck_IPInSubnet(t *testing.T) {
	middleware, err := WithIPCheck("192.168.1.0/24")
	if err != nil {
		t.Fatalf("WithIPCheck returned error: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := middleware(testHandler)

	testCases := []string{
		"192.168.1.1",
		"192.168.1.10",
		"192.168.1.100",
		"192.168.1.255",
	}

	for _, ip := range testCases {
		t.Run(ip, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Real-IP", ip)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("IP %s in subnet should pass, got status %d", ip, w.Code)
			}
		})
	}
}

func TestWithIPCheck_IPNotInSubnet(t *testing.T) {
	middleware, err := WithIPCheck("192.168.1.0/24")
	if err != nil {
		t.Fatalf("WithIPCheck returned error: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	testCases := []string{
		"192.168.2.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
	}

	for _, ip := range testCases {
		t.Run(ip, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Real-IP", ip)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Errorf("IP %s not in subnet should be rejected, got status %d", ip, w.Code)
			}
		})
	}
}

func TestWithIPCheck_SubnetBoundaries(t *testing.T) {
	middleware, err := WithIPCheck("192.168.1.10/32")
	if err != nil {
		t.Fatalf("WithIPCheck returned error: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.10")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Exact IP match should pass, got status %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.11")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Different IP should be rejected, got status %d", w.Code)
	}
}

func TestWithIPCheck_IPv6(t *testing.T) {
	middleware, err := WithIPCheck("2001:db8::/32")
	if err != nil {
		t.Fatalf("WithIPCheck returned error: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "2001:db8::1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("IPv6 in subnet should pass, got status %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "2001:db9::1")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("IPv6 not in subnet should be rejected, got status %d", w.Code)
	}
}
