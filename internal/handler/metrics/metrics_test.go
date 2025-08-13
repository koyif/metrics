package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/pkg/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const failingMetricsName = "failingMetrics"

type MockMetricsRepository struct{}

func (MockMetricsRepository) StoreCounter(metricName string, value int64) error {
	if metricName == failingMetricsName {
		return fmt.Errorf("store error: %s", failingMetricsName)
	}
	return nil
}
func (MockMetricsRepository) StoreGauge(metricName string, value float64) error {
	if metricName == failingMetricsName {
		return fmt.Errorf("store error: %s", failingMetricsName)
	}
	return nil
}
func (MockMetricsRepository) Persist() error {
	return nil
}

func TestStoreHandler_Handle(t *testing.T) {
	var (
		delta int64 = 100
		value       = 100.0
	)
	type given struct {
	}
	type when struct {
		request dto.Metrics
	}
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name  string
		given *given
		when  when
		want  want
	}{
		{
			name: "empty metrics name",
			when: when{
				request: dto.Metrics{
					ID:    "",
					MType: dto.CounterMetricsType,
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name: "unknown metrics type",
			when: when{
				request: dto.Metrics{
					ID:    "test",
					MType: "test",
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "empty delta in counter metrics type",
			when: when{
				request: dto.Metrics{
					ID:    "test",
					MType: dto.CounterMetricsType,
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "empty delta in gauge metrics type",
			when: when{
				request: dto.Metrics{
					ID:    "test",
					MType: dto.GaugeMetricsType,
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "counter storing error",
			when: when{
				request: dto.Metrics{
					ID:    failingMetricsName,
					MType: dto.CounterMetricsType,
					Delta: &delta,
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusInternalServerError,
			},
		},
		{
			name: "gauge storing error",
			when: when{
				request: dto.Metrics{
					ID:    failingMetricsName,
					MType: dto.GaugeMetricsType,
					Value: &value,
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusInternalServerError,
			},
		},
		{
			name: "counter successfully stored",
			when: when{
				request: dto.Metrics{
					ID:    "test",
					MType: dto.CounterMetricsType,
					Delta: &delta,
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
			},
		},
		{
			name: "gauge successfully stored",
			when: when{
				request: dto.Metrics{
					ID:    "test",
					MType: dto.GaugeMetricsType,
					Value: &value,
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
			},
		},
	}

	handler := NewStoreHandler(MockMetricsRepository{}, config.Load())

	mux := http.NewServeMux()
	mux.HandleFunc("/update", handler.Handle)

	server := httptest.NewServer(mux)
	defer server.Close()

	baseURL, err := url.Parse(server.URL)
	assert.NoError(t, err)

	updateURL := baseURL.JoinPath("update").String()

	client := http.Client{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.when.request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, updateURL, bytes.NewReader(body))
			require.NoError(t, err)

			resp, err := client.Do(req)
			if err != nil {
				err := resp.Body.Close()
				if err != nil {
					slog.Error(fmt.Sprintf("Failed to close response body: %s", err))
					return
				}
			}

			require.NoError(t, err)

			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Errorf("error closing response body: %v", err)
				}
			}(resp.Body)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, "application/json")
		})
	}
}
