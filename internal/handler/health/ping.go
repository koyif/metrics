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

// @Summary		Health check
// @Description	Check service health and database connectivity
// @Tags			health
// @Produce		plain
// @Success		200	{string}	string	"pong"
// @Failure		500	{string}	string	"Service Unavailable"
// @Router			/ping [get]
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
