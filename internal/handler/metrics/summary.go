package metrics

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/koyif/metrics/internal/handler"
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

// @Summary		Get all metrics summary
// @Description	Retrieve an HTML page displaying all stored counter and gauge metrics
// @Tags			metrics
// @Produce		html
// @Success		200	{string}	string	"HTML table with all metrics"
// @Failure		500	{string}	string	"Internal Server Error - Template failure"
// @Router			/ [get]
func (h *SummaryHandler) Handle(w http.ResponseWriter, _ *http.Request) {
	res := make(map[string]string)

	for k, v := range h.service.AllGauges() {
		res[k] = strconv.FormatFloat(v, 'f', -1, 64)
	}

	for k, v := range h.service.AllCounters() {
		res[k] = strconv.FormatInt(v, 10)
	}

	tt, err := template.New("summary").Parse(summaryHTML)
	if err != nil {
		handler.InternalServerError(w, err, "failed to parse template")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err = tt.Execute(w, res); err != nil {
		handler.InternalServerError(w, err, "failed to execute template")
		return
	}
}
