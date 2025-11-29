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

// @Summary		Store counter metric (deprecated)
// @Description	Legacy endpoint to store a counter metric via URL path parameters
// @Tags			deprecated
// @Produce		plain
// @Param			metric	path	string	true	"Metric name"
// @Param			value	path	int		true	"Counter value (integer)"
// @Success		200		"OK"
// @Failure		400		{string}	string	"Bad Request - Invalid value format"
// @Failure		404		{string}	string	"Not Found - Empty metric name"
// @Failure		500		{string}	string	"Internal Server Error - Storage failure"
// @Router			/update/counter/{metric}/{value} [post]
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

// @Summary		Get counter metric (deprecated)
// @Description	Legacy endpoint to retrieve a counter metric value via URL path
// @Tags			deprecated
// @Produce		plain
// @Param			metric	path	string	true	"Metric name"
// @Success		200		{string}	string	"Counter value as plain text"
// @Failure		404		{string}	string	"Not Found - Empty metric name or metric not found"
// @Failure		500		{string}	string	"Internal Server Error - Retrieval failure"
// @Router			/value/counter/{metric} [get]
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
