package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/internal/repository/dberror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/pkg/dto"
	"github.com/koyif/metrics/pkg/types"
)

const failingMetricsName = "failingMetrics"

type MockMetricsRepository struct {
	mu                sync.Mutex
	counterCalls      []CounterCall
	gaugeCalls        []GaugeCall
	persistCalls      int
	shouldFailPersist bool
}

type CounterCall struct {
	MetricName string
	Value      int64
}

type GaugeCall struct {
	MetricName string
	Value      float64
}

func NewMockMetricsRepository() *MockMetricsRepository {
	return &MockMetricsRepository{
		counterCalls: make([]CounterCall, 0),
		gaugeCalls:   make([]GaugeCall, 0),
	}
}

func (m *MockMetricsRepository) StoreCounter(metricName string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counterCalls = append(m.counterCalls, CounterCall{
		MetricName: metricName,
		Value:      value,
	})

	if metricName == failingMetricsName {
		return fmt.Errorf("store error: %s", failingMetricsName)
	}
	return nil
}

func (m *MockMetricsRepository) StoreGauge(metricName string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gaugeCalls = append(m.gaugeCalls, GaugeCall{
		MetricName: metricName,
		Value:      value,
	})

	if metricName == failingMetricsName {
		return fmt.Errorf("store error: %s", failingMetricsName)
	}
	return nil
}

func (m *MockMetricsRepository) StoreAll(metrics []models.Metrics) error {
	// Not used in current tests
	return nil
}

func (m *MockMetricsRepository) Persist() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.persistCalls++
	if m.shouldFailPersist {
		return fmt.Errorf("persist failed")
	}
	return nil
}

func (m *MockMetricsRepository) GetCounterCalls() []CounterCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]CounterCall(nil), m.counterCalls...)
}

func (m *MockMetricsRepository) GetGaugeCalls() []GaugeCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]GaugeCall(nil), m.gaugeCalls...)
}

func (m *MockMetricsRepository) GetPersistCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.persistCalls
}

func (m *MockMetricsRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counterCalls = m.counterCalls[:0]
	m.gaugeCalls = m.gaugeCalls[:0]
	m.persistCalls = 0
	m.shouldFailPersist = false
}

func TestGetHandler_Handle(t *testing.T) {
	const (
		counterName = "test_counter"
		gaugeName   = "test_gauge"
	)

	var (
		counterValue int64 = 42
		gaugeValue         = 3.14
	)

	type mockGetterFunc func(string) (interface{}, error)
	type given struct {
		counterFunc mockGetterFunc
		gaugeFunc   mockGetterFunc
	}
	type when struct {
		request     dto.Metrics
		requestBody string
		useRawBody  bool
	}
	type want struct {
		statusCode       int
		responseMetrics  *dto.Metrics
		responseContains string
	}

	tests := []struct {
		name  string
		given given
		when  when
		want  want
	}{
		{
			name: "successfully get counter metric",
			given: given{
				counterFunc: func(name string) (interface{}, error) {
					if name == counterName {
						return counterValue, nil
					}
					return nil, dberror.ErrValueNotFound
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    counterName,
					MType: dto.CounterMetricsType,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				responseMetrics: &dto.Metrics{
					ID:    counterName,
					MType: dto.CounterMetricsType,
					Delta: &counterValue,
				},
			},
		},
		{
			name: "successfully get gauge metric",
			given: given{
				gaugeFunc: func(name string) (interface{}, error) {
					if name == gaugeName {
						return gaugeValue, nil
					}
					return nil, dberror.ErrValueNotFound
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    gaugeName,
					MType: dto.GaugeMetricsType,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				responseMetrics: &dto.Metrics{
					ID:    gaugeName,
					MType: dto.GaugeMetricsType,
					Value: &gaugeValue,
				},
			},
		},
		{
			name: "metric not found",
			given: given{
				counterFunc: func(string) (interface{}, error) {
					return nil, dberror.ErrValueNotFound
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "nonexistent",
					MType: dto.CounterMetricsType,
				},
			},
			want: want{
				statusCode:       http.StatusNotFound,
				responseContains: http.StatusText(http.StatusNotFound),
			},
		},
		{
			name: "malformed JSON request",
			when: when{
				requestBody: `{"invalid": json`,
				useRawBody:  true,
			},
			want: want{
				statusCode:       http.StatusBadRequest,
				responseContains: http.StatusText(http.StatusBadRequest),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter := &mockMetricsGetter{
				counterFunc: tt.given.counterFunc,
				gaugeFunc:   tt.given.gaugeFunc,
			}

			handler := NewGetHandler(mockGetter)

			server := httptest.NewServer(http.HandlerFunc(handler.Handle))
			defer server.Close()

			var body io.Reader
			if tt.when.useRawBody {
				body = strings.NewReader(tt.when.requestBody)
			} else {
				jsonBody, err := json.Marshal(tt.when.request)
				require.NoError(t, err)
				body = bytes.NewReader(jsonBody)
			}

			req, err := http.NewRequest(http.MethodPost, server.URL, body)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.responseMetrics != nil {
				var response dto.Metrics
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, *tt.want.responseMetrics, response)
			}

			if tt.want.responseContains != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Contains(t, string(body), tt.want.responseContains)
			}
		})
	}
}

type mockMetricsGetter struct {
	counterFunc func(string) (interface{}, error)
	gaugeFunc   func(string) (interface{}, error)
}

func (m *mockMetricsGetter) Counter(name string) (int64, error) {
	if m.counterFunc == nil {
		return 0, dberror.ErrValueNotFound
	}
	val, err := m.counterFunc(name)
	if err != nil {
		return 0, err
	}
	return val.(int64), nil
}

func (m *mockMetricsGetter) Gauge(name string) (float64, error) {
	if m.gaugeFunc == nil {
		return 0, dberror.ErrValueNotFound
	}
	val, err := m.gaugeFunc(name)
	if err != nil {
		return 0, err
	}
	return val.(float64), nil
}

func TestStoreHandler_Handle(t *testing.T) {
	var (
		delta      int64 = 100
		value            = 100.0
		zeroDelta  int64 = 0
		zeroValue        = 0.0
		negDelta   int64 = -50
		negValue         = -25.5
		largeDelta int64 = math.MaxInt64
		largeValue       = math.MaxFloat64
	)

	type given struct {
		setupMock func(*MockMetricsRepository)
		config    *config.Config
	}
	type when struct {
		request     dto.Metrics
		requestBody string // for malformed JSON tests
		useRawBody  bool
	}
	type want struct {
		contentType      string
		statusCode       int
		responseContains string
		verifyMock       func(*testing.T, *MockMetricsRepository)
	}

	tests := []struct {
		name  string
		given given
		when  when
		want  want
	}{
		{
			name: "malformed JSON request",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				requestBody: `{"invalid": json}`,
				useRawBody:  true,
			},
			want: want{
				contentType:      "text/plain; charset=utf-8",
				statusCode:       http.StatusBadRequest,
				responseContains: "Bad Request",
			},
		},
		{
			name: "empty metrics name",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "",
					MType: dto.CounterMetricsType,
				},
			},
			want: want{
				contentType:      "text/plain; charset=utf-8",
				statusCode:       http.StatusNotFound,
				responseContains: "Not Found",
			},
		},
		{
			name: "unknown metrics type",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "test",
					MType: "invalid_type",
				},
			},
			want: want{
				contentType:      "text/plain; charset=utf-8",
				statusCode:       http.StatusBadRequest,
				responseContains: "Bad Request",
			},
		},
		{
			name: "empty delta in counter metrics type",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "test",
					MType: dto.CounterMetricsType,
				},
			},
			want: want{
				contentType:      "text/plain; charset=utf-8",
				statusCode:       http.StatusBadRequest,
				responseContains: "Bad Request",
			},
		},
		{
			name: "empty value in gauge metrics type",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "test",
					MType: dto.GaugeMetricsType,
				},
			},
			want: want{
				contentType:      "text/plain; charset=utf-8",
				statusCode:       http.StatusBadRequest,
				responseContains: "Bad Request",
			},
		},
		{
			name: "counter storing error",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    failingMetricsName,
					MType: dto.CounterMetricsType,
					Delta: &delta,
				},
			},
			want: want{
				contentType:      "text/plain; charset=utf-8",
				statusCode:       http.StatusInternalServerError,
				responseContains: "Internal Server Error",
			},
		},
		{
			name: "gauge storing error",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    failingMetricsName,
					MType: dto.GaugeMetricsType,
					Value: &value,
				},
			},
			want: want{
				contentType:      "text/plain; charset=utf-8",
				statusCode:       http.StatusInternalServerError,
				responseContains: "Internal Server Error",
			},
		},
		{
			name: "counter successfully stored",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "test_counter",
					MType: dto.CounterMetricsType,
					Delta: &delta,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				verifyMock: func(t *testing.T, mock *MockMetricsRepository) {
					calls := mock.GetCounterCalls()
					require.Len(t, calls, 1)
					assert.Equal(t, "test_counter", calls[0].MetricName)
					assert.Equal(t, int64(100), calls[0].Value)
				},
			},
		},
		{
			name: "gauge successfully stored",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "test_gauge",
					MType: dto.GaugeMetricsType,
					Value: &value,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				verifyMock: func(t *testing.T, mock *MockMetricsRepository) {
					calls := mock.GetGaugeCalls()
					require.Len(t, calls, 1)
					assert.Equal(t, "test_gauge", calls[0].MetricName)
					assert.Equal(t, 100.0, calls[0].Value)
				},
			},
		},
		{
			name: "counter with zero value",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "zero_counter",
					MType: dto.CounterMetricsType,
					Delta: &zeroDelta,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				verifyMock: func(t *testing.T, mock *MockMetricsRepository) {
					calls := mock.GetCounterCalls()
					require.Len(t, calls, 1)
					assert.Equal(t, "zero_counter", calls[0].MetricName)
					assert.Equal(t, int64(0), calls[0].Value)
				},
			},
		},
		{
			name: "gauge with zero value",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "zero_gauge",
					MType: dto.GaugeMetricsType,
					Value: &zeroValue,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				verifyMock: func(t *testing.T, mock *MockMetricsRepository) {
					calls := mock.GetGaugeCalls()
					require.Len(t, calls, 1)
					assert.Equal(t, "zero_gauge", calls[0].MetricName)
					assert.Equal(t, 0.0, calls[0].Value)
				},
			},
		},
		{
			name: "counter with negative value",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "neg_counter",
					MType: dto.CounterMetricsType,
					Delta: &negDelta,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				verifyMock: func(t *testing.T, mock *MockMetricsRepository) {
					calls := mock.GetCounterCalls()
					require.Len(t, calls, 1)
					assert.Equal(t, "neg_counter", calls[0].MetricName)
					assert.Equal(t, int64(-50), calls[0].Value)
				},
			},
		},
		{
			name: "gauge with negative value",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "neg_gauge",
					MType: dto.GaugeMetricsType,
					Value: &negValue,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				verifyMock: func(t *testing.T, mock *MockMetricsRepository) {
					calls := mock.GetGaugeCalls()
					require.Len(t, calls, 1)
					assert.Equal(t, "neg_gauge", calls[0].MetricName)
					assert.Equal(t, -25.5, calls[0].Value)
				},
			},
		},
		{
			name: "counter with max value",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "max_counter",
					MType: dto.CounterMetricsType,
					Delta: &largeDelta,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				verifyMock: func(t *testing.T, mock *MockMetricsRepository) {
					calls := mock.GetCounterCalls()
					require.Len(t, calls, 1)
					assert.Equal(t, "max_counter", calls[0].MetricName)
					assert.Equal(t, int64(math.MaxInt64), calls[0].Value)
				},
			},
		},
		{
			name: "gauge with max value",
			given: given{
				config: &config.Config{
					Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(300 * time.Second)},
				},
			},
			when: when{
				request: dto.Metrics{
					ID:    "max_gauge",
					MType: dto.GaugeMetricsType,
					Value: &largeValue,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				verifyMock: func(t *testing.T, mock *MockMetricsRepository) {
					calls := mock.GetGaugeCalls()
					require.Len(t, calls, 1)
					assert.Equal(t, "max_gauge", calls[0].MetricName)
					assert.Equal(t, math.MaxFloat64, calls[0].Value)
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mock := NewMockMetricsRepository()
			if tt.given.setupMock != nil {
				tt.given.setupMock(mock)
			}

			handler := NewStoreHandler(mock, tt.given.config, nil)

			mux := http.NewServeMux()
			mux.HandleFunc("/update", handler.Handle)

			server := httptest.NewServer(mux)
			defer server.Close()

			baseURL, err := url.Parse(server.URL)
			assert.NoError(t, err)
			updateURL := baseURL.JoinPath("update").String()

			client := http.Client{}

			// Prepare request body
			var body io.Reader
			if tt.when.useRawBody {
				body = strings.NewReader(tt.when.requestBody)
			} else {
				jsonBody, err := json.Marshal(tt.when.request)
				require.NoError(t, err)
				body = bytes.NewReader(jsonBody)
			}

			req, err := http.NewRequest(http.MethodPost, updateURL, body)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			}

			if tt.want.responseContains != "" {
				responseBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Contains(t, string(responseBody), tt.want.responseContains)
			}

			// Verify mock interactions
			if tt.want.verifyMock != nil {
				tt.want.verifyMock(t, mock)
			}
		})
	}
}

func TestStoreHandler_PersistBehavior(t *testing.T) {
	delta := int64(100)

	tests := []struct {
		name          string
		storeInterval time.Duration
		expectPersist bool
	}{
		{
			name:          "persist when store interval is zero",
			storeInterval: 0,
			expectPersist: true,
		},
		{
			name:          "no persist when store interval is non-zero",
			storeInterval: 30 * time.Second,
			expectPersist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockMetricsRepository()

			cfg := &config.Config{
				Storage: config.StorageConfig{StoreInterval: types.DurationInSeconds(tt.storeInterval)},
			}

			handler := NewStoreHandler(mock, cfg, nil)

			mux := http.NewServeMux()
			mux.HandleFunc("/update", handler.Handle)

			server := httptest.NewServer(mux)
			defer server.Close()

			baseURL, err := url.Parse(server.URL)
			require.NoError(t, err)
			updateURL := baseURL.JoinPath("update").String()

			request := dto.Metrics{
				ID:    "test_persist",
				MType: dto.CounterMetricsType,
				Delta: &delta,
			}

			body, err := json.Marshal(request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, updateURL, bytes.NewReader(body))
			require.NoError(t, err)

			client := http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Verify persist behavior
			persistCalls := mock.GetPersistCalls()
			if tt.expectPersist {
				assert.Equal(t, 1, persistCalls, "Expected Persist() to be called once")
			} else {
				assert.Equal(t, 0, persistCalls, "Expected Persist() not to be called")
			}
		})
	}
}
