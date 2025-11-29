package metrics

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/koyif/metrics/internal/audit"
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

// StoreHandler handles HTTP requests for storing a single metric.
// It processes POST requests at /update/ with JSON body containing metric data.
type StoreHandler struct {
	service      metricsStorer
	cfg          *config.Config
	auditManager *audit.Manager
}

// StoreAllHandler handles HTTP requests for batch storing multiple metrics.
// It processes POST requests at /updates/ with JSON array of metrics.
type StoreAllHandler struct {
	service      metricsStorer
	cfg          *config.Config
	auditManager *audit.Manager
}

// GetHandler handles HTTP requests for retrieving metric values.
// It processes POST requests at /value/ with JSON body specifying the metric to retrieve.
type GetHandler struct {
	service metricsGetter
}

// NewStoreHandler creates a new handler for single metric storage.
// The auditManager can be nil if auditing is not enabled.
func NewStoreHandler(service metricsStorer, cfg *config.Config, auditManager *audit.Manager) *StoreHandler {
	return &StoreHandler{
		service:      service,
		cfg:          cfg,
		auditManager: auditManager,
	}
}

// NewStoreAllHandler creates a new handler for batch metric storage.
// The auditManager can be nil if auditing is not enabled.
func NewStoreAllHandler(service metricsStorer, cfg *config.Config, auditManager *audit.Manager) *StoreAllHandler {
	return &StoreAllHandler{
		service:      service,
		cfg:          cfg,
		auditManager: auditManager,
	}
}

// NewGetHandler creates a new handler for metric retrieval.
func NewGetHandler(service metricsGetter) *GetHandler {
	return &GetHandler{
		service: service,
	}
}

// @Summary		Store a single metric
// @Description	Store a single counter or gauge metric with validation and optional persistence
// @Tags			metrics
// @Accept			json
// @Produce		plain
// @Param			metric	body	dto.Metrics	true	"Metric data (counter with delta or gauge with value)"
// @Success		200		"OK"
// @Failure		400		{string}	string	"Bad Request - Invalid JSON or unknown metric type"
// @Failure		404		{string}	string	"Not Found - Empty metric ID"
// @Failure		500		{string}	string	"Internal Server Error - Storage failure"
// @Router			/update/ [post]
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

	if sh.cfg.Storage.StoreInterval.Value() == 0 {
		if err := sh.service.Persist(); err != nil {
			logger.Log.Warn(failedToPersistMetricsErrorMessage, logger.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}
	}

	sendAuditEvent(sh.auditManager, []string{m.ID}, getClientIP(r))

	w.WriteHeader(http.StatusOK)
}

// @Summary		Store multiple metrics in batch
// @Description	Store an array of metrics (counters and gauges) in a single batch operation
// @Tags			metrics
// @Accept			json
// @Produce		plain
// @Param			metrics	body	[]dto.Metrics	true	"Array of metrics to store"
// @Success		200		"OK"
// @Failure		400		{string}	string	"Bad Request - Invalid JSON format"
// @Failure		404		{string}	string	"Not Found - Empty metric ID or empty array"
// @Failure		500		{string}	string	"Internal Server Error - Storage failure"
// @Router			/updates/ [post]
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

	metricNames := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		metricNames = append(metricNames, metric.ID)
	}
	sendAuditEvent(sh.auditManager, metricNames, getClientIP(r))

	w.WriteHeader(http.StatusOK)
}

// @Summary		Get a metric value
// @Description	Retrieve the current value of a counter or gauge metric
// @Tags			metrics
// @Accept			json
// @Produce		json
// @Param			metric	body		dto.Metrics	true	"Metric identifier (id and type required)"
// @Success		200		{object}	dto.Metrics	"Metric with current value (delta for counter, value for gauge)"
// @Failure		400		{string}	string		"Bad Request - Invalid JSON or unknown metric type"
// @Failure		404		{string}	string		"Not Found - Empty metric ID or metric not found"
// @Failure		500		{string}	string		"Internal Server Error - Retrieval failure"
// @Router			/value/ [post]
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

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

func sendAuditEvent(auditManager *audit.Manager, metricNames []string, ipAddress string) {
	if auditManager == nil || !auditManager.IsEnabled() {
		return
	}

	event := models.AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: ipAddress,
	}

	auditManager.NotifyAll(event)
}
