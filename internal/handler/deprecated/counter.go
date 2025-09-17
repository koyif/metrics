package deprecated

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/koyif/metrics/internal/repository/dberror"
	"github.com/koyif/metrics/pkg/logger"
)

type counterStorer interface {
	StoreCounter(metricName string, value int64) error
}

type counterGetter interface {
	Counter(metricName string) (int64, error)
}

type CountersPostHandler struct {
	service counterStorer
}

type CountersGetHandler struct {
	service counterGetter
}

func NewCountersPostHandler(service counterStorer) *CountersPostHandler {
	return &CountersPostHandler{
		service: service,
	}
}

func NewCountersGetHandler(service counterGetter) *CountersGetHandler {
	return &CountersGetHandler{
		service: service,
	}
}

func (ch CountersPostHandler) Handle(w http.ResponseWriter, r *http.Request) {
	mn := r.PathValue("metric")
	value := r.PathValue("value")
	if mn == "" || value == "" {
		logger.Log.Warn(metricIDEmptyErrorMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	if v, err := strconv.ParseInt(value, 10, 64); err != nil {
		logger.Log.Warn(incorrectValueFormatMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	} else if err := ch.service.StoreCounter(mn, v); err != nil {
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

func (h CountersGetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	mn := r.PathValue("metric")
	if mn == "" {
		logger.Log.Warn(metricIDEmptyErrorMessage, logger.String("URI", r.RequestURI))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	value, err := h.service.Counter(mn)
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
	_, _ = w.Write([]byte(strconv.FormatInt(value, 10)))
}
