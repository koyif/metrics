package metrics

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/internal/repository/dberror"

	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/pkg/dto"
	"github.com/koyif/metrics/pkg/logger"
)

const (
	incorrectJSONFormatMessage         = "incorrect JSON format"
	metricIDEmptyErrorMessage          = "metric ID cannot be empty"
	unknownMetricTypeMessage           = "unknown metric type"
	emptyMetricsErrorMessage           = "metrics array cannot be empty"
	failedToPersistMetricsErrorMessage = "failed to persist metrics"
	valueNotFoundErrorMessage          = "value not found in storage"
	failedToGetMetricValueErrorMessage = "failed to get metric value"
	failedToEncodeErrorMessage         = "failed to encode response"
	nilValueErrorMessage               = "incorrect value format: nil"
)

type metricsStorer interface {
	StoreCounter(metricName string, value int64) error
	StoreGauge(metricName string, value float64) error
	StoreAll(metrics []models.Metrics) error
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

type StoreAllHandler struct {
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

func NewStoreAllHandler(service metricsStorer, cfg *config.Config) *StoreAllHandler {
	return &StoreAllHandler{
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
		logger.Log.Warn(incorrectJSONFormatMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if m.ID == "" {
		logger.Log.Warn(metricIDEmptyErrorMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	switch m.MType {
	case dto.CounterMetricsType:
		sh.handleCounter(w, r, m.ID, m.Delta)
	case dto.GaugeMetricsType:
		sh.handleGauge(w, r, m.ID, m.Value)
	default:
		logger.Log.Warn(unknownMetricTypeMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if sh.cfg.Storage.StoreInterval == 0 {
		if err := sh.service.Persist(); err != nil {
			logger.Log.Warn(failedToPersistMetricsErrorMessage, logger.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (sh StoreAllHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var m []dto.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log.Warn(incorrectJSONFormatMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if len(m) == 0 {
		logger.Log.Warn(emptyMetricsErrorMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	metrics := make([]models.Metrics, 0, len(m))

	for _, metric := range m {
		if metric.ID == "" {
			logger.Log.Warn(metricIDEmptyErrorMessage, logger.String("URI", r.RequestURI))
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

			return
		}

		metrics = append(metrics, models.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Delta: metric.Delta,
			Value: metric.Value,
		})
	}

	if err := sh.service.StoreAll(metrics); err != nil {
		logger.Log.Warn(failedToPersistMetricsErrorMessage, logger.Error(err))
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (gh GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var m dto.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log.Warn(incorrectJSONFormatMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if m.ID == "" {
		logger.Log.Warn(metricIDEmptyErrorMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

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
		logger.Log.Warn("unknown metric type", logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if valErr != nil && errors.Is(valErr, dberror.ErrValueNotFound) {
		logger.Log.Warn(valueNotFoundErrorMessage, logger.String("URI", r.RequestURI), logger.String("ID", m.ID))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	} else if valErr != nil {
		logger.Log.Warn(failedToGetMetricValueErrorMessage, logger.Error(valErr))
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(m); err != nil {
		logger.Log.Warn(failedToEncodeErrorMessage, logger.Error(err))
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}
}

func (sh StoreHandler) handleCounter(w http.ResponseWriter, r *http.Request, metricName string, value *int64) {
	if value == nil {
		logger.Log.Warn(nilValueErrorMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if err := sh.service.StoreCounter(metricName, *value); err != nil {
		logger.Log.Warn(failedToPersistMetricsErrorMessage, logger.Error(err))
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}
}

func (sh StoreHandler) handleGauge(w http.ResponseWriter, r *http.Request, metricName string, value *float64) {
	if value == nil {
		logger.Log.Warn(nilValueErrorMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if err := sh.service.StoreGauge(metricName, *value); err != nil {
		logger.Log.Warn(failedToPersistMetricsErrorMessage, logger.Error(err))
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}
}
