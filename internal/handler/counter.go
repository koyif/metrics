package handler

import (
	"errors"
	"github.com/koyif/metrics/internal/repository"
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
		InvalidMethodError(w, r)
		return
	}

	metricName := r.PathValue("metric")
	value := r.PathValue("value")
	if metricName == "" || value == "" {
		MetricNameNotPresentError(w, r)
		return
	}

	metricValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		IncorrectValueError(w, value)
		return
	}

	if err := ch.service.StoreCounter(metricName, metricValue); err != nil {
		StoreError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h CountersGetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		InvalidMethodError(w, r)
		return
	}

	metricName := r.PathValue("metric")
	if metricName == "" {
		MetricNameNotPresentError(w, r)
		return
	}

	value, err := h.service.Counter(metricName)
	if err != nil && errors.Is(err, repository.ErrValueNotFound) {
		ValueNotFoundError(w, metricName)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatInt(value, 10)))
}
