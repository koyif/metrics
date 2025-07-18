package handler

import (
	"errors"
	"fmt"
	"github.com/koyif/metrics/internal/repository"
	"log/slog"
	"net/http"
	"strconv"
)

type gaugeStorer interface {
	StoreGauge(metricName string, value float64) error
}

type gaugeGetter interface {
	Gauge(metricName string) (float64, error)
}

type GaugesPostHandler struct {
	service gaugeStorer
}

type GaugesGetHandler struct {
	service gaugeGetter
}

func NewGaugesPostHandler(service gaugeStorer) *GaugesPostHandler {
	return &GaugesPostHandler{
		service: service,
	}
}

func NewGaugesGetHandler(service gaugeGetter) *GaugesGetHandler {
	return &GaugesGetHandler{
		service: service,
	}
}

func (h GaugesPostHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		invalidMethodError(w, r)
		return
	}

	metricName := r.PathValue("metric")
	value := r.PathValue("value")
	if metricName == "" || value == "" {
		metricNameNotPresentError(w, r)
		return
	}

	metricValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		incorrectValueError(w, value)
		return
	}

	if err := h.service.StoreGauge(metricName, metricValue); err != nil {
		storeError(w, err)
		return
	}

	slog.Debug(fmt.Sprintf("stored: %s: %f", metricName, metricValue))

	w.WriteHeader(http.StatusOK)
}

func (h GaugesGetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		invalidMethodError(w, r)
		return
	}

	metricName := r.PathValue("metric")
	if metricName == "" {
		metricNameNotPresentError(w, r)
		return
	}

	value, err := h.service.Gauge(metricName)
	if err != nil && errors.Is(err, repository.ErrValueNotFound) {
		valueNotFoundError(w, metricName)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
}
