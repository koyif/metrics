package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

type GaugesRepository interface {
	StoreGauge(metricName string, value float64) error
}

type GaugesHandler struct {
	repository GaugesRepository
}

func NewGaugesHandler(repository GaugesRepository) *GaugesHandler {
	return &GaugesHandler{
		repository: repository,
	}
}

func (h GaugesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	const op = "GaugesHandler.Handle"

	if r.Method != http.MethodPost {
		invalidMethodError(w, r, op)
		return
	}

	metricName := r.PathValue("metric")
	if metricName == "" {
		metricNameNotPresentError(w, r, op)
		return
	}

	value := r.PathValue("value")
	if value == "" {
		valueNotPresentError(w, r, op)
		return
	}

	metricValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		incorrectValueError(w, op, value)
		return
	}

	if err := h.repository.StoreGauge(metricName, metricValue); err != nil {
		storeError(w, op, err)
		return
	}

	slog.Info(fmt.Sprintf("%s: stored: %s: %f", op, metricName, metricValue))

	w.WriteHeader(http.StatusOK)
}
