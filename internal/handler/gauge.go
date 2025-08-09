package handler

import (
	"errors"
	"github.com/koyif/metrics/internal/repository"
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
		InvalidMethodError(w, r)
		return
	}

	metricName := r.PathValue("metric")
	value := r.PathValue("value")
	if metricName == "" || value == "" {
		MetricNameNotPresentError(w, r)
		return
	}

	metricValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		IncorrectValueError(w, value)
		return
	}

	if err := h.service.StoreGauge(metricName, metricValue); err != nil {
		StoreError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h GaugesGetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		InvalidMethodError(w, r)
		return
	}

	metricName := r.PathValue("metric")
	if metricName == "" {
		MetricNameNotPresentError(w, r)
		return
	}

	value, err := h.service.Gauge(metricName)
	if err != nil && errors.Is(err, repository.ErrValueNotFound) {
		ValueNotFoundError(w, metricName)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
}
