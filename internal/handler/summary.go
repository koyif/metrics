package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
)

const summaryHTML = `<html>
	<body>
		{{- $size := len . }}
		{{ if eq $size 0 -}}
		<span>Database is empty</span>
		{{- else -}}
		<table>
			{{- range $name, $value := . -}}
				<tr><td>{{ $name }}</td><td>{{ $value }}</td></tr>
			{{- end -}}
		</table>
		{{- end -}}
	</body>
</html>`

type summaryGetter interface {
	AllCounters() map[string]int64
	AllGauges() map[string]float64
}
type SummaryHandler struct {
	service summaryGetter
}

func NewSummaryHandler(service summaryGetter) *SummaryHandler {
	return &SummaryHandler{
		service: service,
	}
}

func (h *SummaryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	const op = "SummaryHandler.Handle"
	if r.Method != http.MethodGet {
		invalidMethodError(w, r, op)
		return
	}

	res := make(map[string]string)

	for k, v := range h.service.AllGauges() {
		res[k] = strconv.FormatFloat(v, 'f', -1, 64)
	}

	for k, v := range h.service.AllCounters() {
		res[k] = strconv.FormatInt(v, 10)
	}

	tt, err := template.New("summary").Parse(summaryHTML)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	for k, v := range res {
		slog.Info(k, " ", v)
	}

	if err = tt.Execute(w, res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
