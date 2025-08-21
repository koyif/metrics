package metrics

import (
	"encoding/json"
	"errors"
	"fmt"

	"net/http"

	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/pkg/dto"
)

type metricsStorer interface {
	StoreCounter(metricName string, value int64) error
	StoreGauge(metricName string, value float64) error
	Persist() error
}

type metricsGetter interface {
	Counter(metricName string) (int64, error)
	Gauge(metricName string) (float64, error)
}

type StoreHandler struct {
	service metricsStorer
	cfg     *config.Config
}

type GetHandler struct {
	service metricsGetter
}

func NewStoreHandler(service metricsStorer, cfg *config.Config) *StoreHandler {
	return &StoreHandler{
		service: service,
		cfg:     cfg,
	}
}

func NewGetHandler(service metricsGetter) *GetHandler {
	return &GetHandler{
		service: service,
	}
}

func (sh StoreHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var m dto.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		handler.BadRequest(w, r.RequestURI, "incorrect JSON format")
		return
	}

	if m.ID == "" {
		handler.NotFound(w, r, "")
		return
	}

	switch m.MType {
	case dto.CounterMetricsType:
		sh.handleCounter(w, m.ID, m.Delta)
	case dto.GaugeMetricsType:
		sh.handleGauge(w, m.ID, m.Value)
	default:
		handler.UnknownMetricTypeHandler(w, r)
		return
	}

	if sh.cfg.Storage.StoreInterval == 0 {
		err := sh.service.Persist()
		if err != nil {
			handler.BadRequest(w, r.RequestURI, "failed to persist metrics")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (gh GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var m dto.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		handler.BadRequest(w, r.RequestURI, "incorrect JSON format")
		return
	}

	if m.ID == "" {
		handler.NotFound(w, r, "")
		return
	}

	var valErr error
	switch m.MType {
	case dto.CounterMetricsType:
		del, err := gh.service.Counter(m.ID)
		valErr = err
		m.Delta = &del
	case dto.GaugeMetricsType:
		val, err := gh.service.Gauge(m.ID)
		valErr = err
		m.Value = &val
	default:
		handler.UnknownMetricTypeHandler(w, r)
		return
	}

	if valErr != nil && errors.Is(valErr, repository.ErrValueNotFound) {
		handler.NotFound(w, r, fmt.Sprintf("value not found in storage: %s", m.ID))
		return
	} else if valErr != nil {
		handler.InternalServerError(w, valErr, "failed to get metric value")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(m); err != nil {
		handler.InternalServerError(w, err, "failed to encode response")
	}
}

func (sh StoreHandler) handleCounter(w http.ResponseWriter, metricName string, value *int64) {
	if value == nil {
		handler.BadRequest(w, "", "incorrect value format: nil")
		return
	}

	err := sh.service.StoreCounter(metricName, *value)
	if err != nil {
		handler.InternalServerError(w, err, "failed to store metric")
		return
	}
}

func (sh StoreHandler) handleGauge(w http.ResponseWriter, metricName string, value *float64) {
	if value == nil {
		handler.BadRequest(w, "", "incorrect value format: nil")
		return
	}

	err := sh.service.StoreGauge(metricName, *value)
	if err != nil {
		handler.InternalServerError(w, err, "failed to store metric")
		return
	}
}
