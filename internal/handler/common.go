package handler

import (
	"fmt"
	"log/slog"
	"net/http"
)

func storeError(w http.ResponseWriter, op string, err error) {
	slog.Error(fmt.Sprintf("%s: failed to store counter: %s", op, err.Error()))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func incorrectValueError(w http.ResponseWriter, op string, value string) {
	slog.Error(fmt.Sprintf("%s: incorrect value: %s", op, value))
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func valueNotPresentError(w http.ResponseWriter, r *http.Request, op string) {
	slog.Error(fmt.Sprintf("%s: value not found in the path: %s", op, r.URL.Path))
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func metricNameNotPresentError(w http.ResponseWriter, r *http.Request, op string) {
	slog.Error(fmt.Sprintf("%s: metric name not found in the path: %s", op, r.URL.Path))
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func invalidMethodError(w http.ResponseWriter, r *http.Request, op string) {
	slog.Error(fmt.Sprintf("%s: invalid method: %s", op, r.Method))
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}
