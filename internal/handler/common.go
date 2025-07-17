package handler

import (
	"fmt"
	"log/slog"
	"net/http"
)

func UnknownMetricTypeHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func storeError(w http.ResponseWriter, err error) {
	slog.Error(fmt.Sprintf("failed to store counter: %s", err.Error()))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func incorrectValueError(w http.ResponseWriter, value string) {
	slog.Error(fmt.Sprintf("incorrect value: %s", value))
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func valueNotFoundError(w http.ResponseWriter, metricName string) {
	slog.Error(fmt.Sprintf("value not found in storage: %s", metricName))
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func metricNameNotPresentError(w http.ResponseWriter, r *http.Request) {
	slog.Error(fmt.Sprintf("metric name not found in the path: %s", r.URL.Path))
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func invalidMethodError(w http.ResponseWriter, r *http.Request) {
	slog.Error(fmt.Sprintf("invalid method: %s", r.Method))
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}
