package handler

import (
	"errors"
	"fmt"
	"github.com/koyif/metrics/internal/repository"
	"log/slog"
	"net/http"
	"strconv"
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
	if r.Method != http.MethodPost {
		invalidMethodError(w, r)
		return
	}

	metricName := r.PathValue("metric")
	value := r.PathValue("value")
	if metricName == "" || value == "" {
		metricNameNotPresentError(w, r)
		return
	}

	metricValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		incorrectValueError(w, value)
		return
	}

	if err := ch.service.StoreCounter(metricName, metricValue); err != nil {
		storeError(w, err)
		return
	}
	slog.Debug(fmt.Sprintf("stored: %s: %d", metricName, metricValue))

	w.WriteHeader(http.StatusOK)
}

func (h CountersGetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		invalidMethodError(w, r)
		return
	}

	metricName := r.PathValue("metric")
	if metricName == "" {
		metricNameNotPresentError(w, r)
		return
	}

	value, err := h.service.Counter(metricName)
	if err != nil && errors.Is(err, repository.ErrValueNotFound) {
		valueNotFoundError(w, metricName)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatInt(value, 10)))
}
