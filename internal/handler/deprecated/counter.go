package deprecated

import (
	"errors"
	"fmt"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/repository/dberror"
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
	mn := r.PathValue("metric")
	value := r.PathValue("value")
	if mn == "" || value == "" {
		handler.NotFound(w, r, "")
		return
	}

	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		handler.BadRequest(w, r.RequestURI, fmt.Sprintf("incorrect value format: %s", value))
		return
	}

	if err := ch.service.StoreCounter(mn, v); err != nil {
		handler.InternalServerError(w, err, "failed to store metric")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h CountersGetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	mn := r.PathValue("metric")
	if mn == "" {
		handler.NotFound(w, r, "")
		return
	}

	value, err := h.service.Counter(mn)
	if err != nil && errors.Is(err, dberror.ErrValueNotFound) {
		handler.NotFound(w, r, "value not found in storage")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatInt(value, 10)))
}
