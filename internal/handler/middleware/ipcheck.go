package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/koyif/metrics/pkg/logger"
)

// WithIPCheck creates middleware that validates client IP against trusted subnet.
// Returns error if trustedSubnet is invalid CIDR notation.
func WithIPCheck(trustedSubnet string) (func(http.Handler) http.Handler, error) {
	if trustedSubnet == "" {
		return func(h http.Handler) http.Handler {
			return h
		}, nil
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR notation for trusted subnet: %w", err)
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				logger.Log.Warn(
					"missing X-Real-IP header",
					logger.String("URI", r.RequestURI),
					logger.String("RemoteAddr", r.RemoteAddr),
				)
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(clientIP)
			if ip == nil {
				logger.Log.Warn(
					"invalid IP address in X-Real-IP header",
					logger.String("X-Real-IP", clientIP),
					logger.String("URI", r.RequestURI),
				)
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			if !ipNet.Contains(ip) {
				logger.Log.Warn(
					"IP address not in trusted subnet",
					logger.String("X-Real-IP", clientIP),
					logger.String("TrustedSubnet", trustedSubnet),
					logger.String("URI", r.RequestURI),
				)
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			logger.Log.Debug(
				"IP check passed",
				logger.String("X-Real-IP", clientIP),
				logger.String("TrustedSubnet", trustedSubnet),
			)

			h.ServeHTTP(w, r)
		})
	}, nil
}
