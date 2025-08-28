package deprecated

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/koyif/metrics/internal/repository/dberror"
	"github.com/koyif/metrics/pkg/logger"
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
		logger.Log.Warn(metricIDEmptyErrorMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logger.Log.Warn(incorrectValueFormatMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if err := h.service.StoreGauge(mn, v); err != nil {
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

func (h GaugesGetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	mn := r.PathValue("metric")
	if mn == "" {
		logger.Log.Warn(metricIDEmptyErrorMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	value, err := h.service.Gauge(mn)
	if err != nil && errors.Is(err, dberror.ErrValueNotFound) {
		logger.Log.Warn(valueNotFoundErrorMessage, logger.String("URI", r.RequestURI), logger.String("ID", mn))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	} else if err != nil {
		logger.Log.Warn(failedToGetMetricValueErrorMessage, logger.Error(err))
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
}
