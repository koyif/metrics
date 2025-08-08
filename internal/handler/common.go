package handler

import (
	"github.com/koyif/metrics/internal/app/logger"
	"net/http"
)

func UnknownMetricTypeHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(
		w,
		http.StatusText(http.StatusBadRequest),
		http.StatusBadRequest,
	)
}

func StoreError(w http.ResponseWriter, err error) {
	logger.Log.Warn(
		"failed to store metric",
		logger.Error(err),
	)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

func IncorrectValueError(w http.ResponseWriter, value string) {
	logger.Log.Warn(
		"incorrect value",
		logger.String("value", value),
	)
	http.Error(
		w,
		http.StatusText(http.StatusBadRequest),
		http.StatusBadRequest,
	)
}

func ValueNotFoundError(w http.ResponseWriter, metricName string) {
	logger.Log.Warn(
		"value not found in storage",
		logger.String("metricName", metricName),
	)
	http.Error(
		w,
		http.StatusText(http.StatusNotFound),
		http.StatusNotFound,
	)
}

func MetricNameNotPresentError(w http.ResponseWriter, r *http.Request) {
	logger.Log.Warn(
		"metric name not present",
		logger.String("URI", r.RequestURI),
	)
	http.Error(
		w,
		http.StatusText(http.StatusNotFound),
		http.StatusNotFound,
	)
}

func InvalidMethodError(w http.ResponseWriter, r *http.Request) {
	logger.Log.Warn(
		"invalid method",
		logger.String("URI", r.RequestURI),
		logger.String("method", r.Method),
	)
	http.Error(
		w,
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed,
	)
}

func IncorrectJsonFormatError(w http.ResponseWriter, r *http.Request) {
	logger.Log.Warn(
		"incorrect JSON format",
		logger.String("URI", r.RequestURI),
	)
	http.Error(
		w,
		http.StatusText(http.StatusBadRequest),
		http.StatusBadRequest,
	)
}

func MarshallingError(w http.ResponseWriter, err error) {
	logger.Log.Warn(
		"error marshalling to JSON",
		logger.Error(err),
	)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}
