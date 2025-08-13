package handler

import (
	"github.com/koyif/metrics/internal/app/logger"
	"net/http"
)

func UnknownMetricTypeHandler(w http.ResponseWriter, r *http.Request) {
	BadRequest(w, r.RequestURI, "unknown metric type")
}

func InternalServerError(w http.ResponseWriter, err error, m string) {
	logger.Log.Warn(m, logger.Error(err))
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

func NotFound(w http.ResponseWriter, r *http.Request, m string) {
	logger.Log.Warn(m, logger.String("URI", r.RequestURI))
	http.Error(
		w,
		http.StatusText(http.StatusNotFound),
		http.StatusNotFound,
	)
}

func BadRequest(w http.ResponseWriter, uri, m string) {
	logger.Log.Warn(m, logger.String("URI", uri))
	http.Error(
		w,
		http.StatusText(http.StatusBadRequest),
		http.StatusBadRequest,
	)
}
