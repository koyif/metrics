package health

import (
	"context"
	"net/http"

	"github.com/koyif/metrics/pkg/logger"
)

type pingService interface {
	Ping(ctx context.Context) error
}

// PingHandler handles HTTP health check requests.
// It processes GET requests at /ping to verify service availability and database connectivity.
type PingHandler struct {
	service pingService
}

// NewPingHandler creates a new health check handler.
func NewPingHandler(service pingService) *PingHandler {
	return &PingHandler{
		service: service,
	}
}

// Handle processes a health check request.
// It performs a database ping (if database storage is configured) and returns the service status.
//
// Returns:
//   - 200 OK: Service is healthy and database is reachable (response body: "pong")
//   - 500 Internal Server Error: Database is unreachable or service is unavailable
func (h *PingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.service.Ping(ctx); err != nil {
		http.Error(w, "Service Unavailable", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("pong"))
	if err != nil {
		logger.Log.Error("error writing response", logger.Error(err))
	}
}
