package health

import (
	"context"
	"net/http"

	"github.com/koyif/metrics/pkg/logger"
)

type pingService interface {
	Ping(ctx context.Context) error
}

type PingHandler struct {
	service pingService
}

func NewPingHandler(service pingService) *PingHandler {
	return &PingHandler{
		service: service,
	}
}

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
