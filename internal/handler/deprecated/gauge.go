package deprecated

import (
	"errors"
	"fmt"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/repository/dberror"
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
	mn := r.PathValue("metric")
	value := r.PathValue("value")
	if mn == "" || value == "" {
		handler.NotFound(w, r, "metric name not present")
		return
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		handler.BadRequest(w, r.RequestURI, fmt.Sprintf("incorrect value format: %s", value))
		return
	}

	if err := h.service.StoreGauge(mn, v); err != nil {
		handler.InternalServerError(w, err, "failed to store metric")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h GaugesGetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	mn := r.PathValue("metric")
	if mn == "" {
		handler.NotFound(w, r, "")
		return
	}

	value, err := h.service.Gauge(mn)
	if err != nil && errors.Is(err, dberror.ErrValueNotFound) {
		handler.NotFound(w, r, "value not found in storage")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
}
