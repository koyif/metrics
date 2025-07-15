package handler

import (
	"net/http"
	"strconv"
)

type CountersRepository interface {
	StoreCounter(metricName string, value int64) error
}

type CountersHandler struct {
	repository CountersRepository
}

func NewCountersHandler(repository CountersRepository) *CountersHandler {
	return &CountersHandler{
		repository: repository,
	}
}

func (ch CountersHandler) Handle(w http.ResponseWriter, r *http.Request) {
	const op = "CountersHandler.Handle"

	if r.Method != http.MethodPost {
		invalidMethodError(w, r, op)
		return
	}

	metricName := r.PathValue("metric")
	if metricName == "" {
		metricNameNotPresentError(w, r, op)
		return
	}

	value := r.PathValue("value")
	if value == "" {
		valueNotPresentError(w, r, op)
		return
	}

	metricValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		incorrectValueError(w, op, value)
		return
	}

	if err := ch.repository.StoreCounter(metricName, metricValue); err != nil {
		storeError(w, op, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
