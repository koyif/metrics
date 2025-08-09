package metrics

import (
	"encoding/json"
	"errors"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/pkg/dto"
	"net/http"
)

type metricsStorer interface {
	StoreCounter(metricName string, value int64) error
	StoreGauge(metricName string, value float64) error
}

type metricsGetter interface {
	Counter(metricName string) (int64, error)
	Gauge(metricName string) (float64, error)
}

type StoreHandler struct {
	service metricsStorer
}

type GetHandler struct {
	service metricsGetter
}

func NewStoreHandler(service metricsStorer) *StoreHandler {
	return &StoreHandler{
		service: service,
	}
}

func NewGetHandler(service metricsGetter) *GetHandler {
	return &GetHandler{
		service: service,
	}
}

func (sh StoreHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handler.InvalidMethodError(w, r)
		return
	}

	var metrics dto.Metrics

	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		handler.IncorrectJSONFormatError(w, r)
		return
	}

	if metrics.ID == "" {
		handler.MetricNameNotPresentError(w, r)
		return
	}

	switch metrics.MType {
	case dto.CounterMetricsType:
		sh.handleCounter(w, metrics.ID, metrics.Delta)
	case dto.GaugeMetricsType:
		sh.handleGauge(w, metrics.ID, metrics.Value)
	default:
		handler.UnknownMetricTypeHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (gh GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handler.InvalidMethodError(w, r)
		return
	}

	var metrics dto.Metrics

	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		handler.IncorrectJSONFormatError(w, r)
		return
	}

	if metrics.ID == "" {
		handler.MetricNameNotPresentError(w, r)
		return
	}

	var valErr error
	switch metrics.MType {
	case dto.CounterMetricsType:
		del, err := gh.service.Counter(metrics.ID)
		valErr = err
		metrics.Delta = &del
	case dto.GaugeMetricsType:
		val, err := gh.service.Gauge(metrics.ID)
		valErr = err
		metrics.Value = &val
	default:
		handler.UnknownMetricTypeHandler(w, r)
		return
	}

	if valErr != nil && errors.Is(valErr, repository.ErrValueNotFound) {
		handler.ValueNotFoundError(w, metrics.ID)
		return
	}

	response, err := json.Marshal(metrics)
	if err != nil {
		handler.MarshallingError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func (sh StoreHandler) handleCounter(w http.ResponseWriter, metricName string, value *int64) {
	if value == nil {
		handler.IncorrectValueError(w, "nil")
		return
	}

	err := sh.service.StoreCounter(metricName, *value)
	if err != nil {
		handler.StoreError(w, err)
		return
	}
}

func (sh StoreHandler) handleGauge(w http.ResponseWriter, metricName string, value *float64) {
	if value == nil {
		handler.IncorrectValueError(w, "nil")
		return
	}

	err := sh.service.StoreGauge(metricName, *value)
	if err != nil {
		handler.StoreError(w, err)
		return
	}
}
