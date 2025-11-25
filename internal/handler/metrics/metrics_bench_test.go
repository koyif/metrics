package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/pkg/dto"
)

// mockMetricsService is a mock implementation for benchmarking
type mockMetricsService struct {
	mu       sync.RWMutex
	counters map[string]int64
	gauges   map[string]float64
}

func newMockMetricsService() *mockMetricsService {
	return &mockMetricsService{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (m *mockMetricsService) StoreCounter(metricName string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[metricName] += value
	return nil
}

func (m *mockMetricsService) StoreGauge(metricName string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[metricName] = value
	return nil
}

func (m *mockMetricsService) StoreAll(metrics []models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				m.gauges[metric.ID] = *metric.Value
			}
		case models.Counter:
			if metric.Delta != nil {
				m.counters[metric.ID] += *metric.Delta
			}
		}
	}
	return nil
}

func (m *mockMetricsService) Counter(metricName string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if v, ok := m.counters[metricName]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *mockMetricsService) Gauge(metricName string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if v, ok := m.gauges[metricName]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *mockMetricsService) Persist() error {
	return nil
}

func (m *mockMetricsService) Ping(ctx context.Context) error {
	return nil
}

// BenchmarkStoreHandler_Counter measures performance of storing counter metrics
func BenchmarkStoreHandler_Counter(b *testing.B) {
	service := newMockMetricsService()
	cfg := &config.Config{}
	handler := NewStoreHandler(service, cfg, nil)

	delta := int64(1)
	metric := dto.Metrics{
		ID:    "test_counter",
		MType: dto.CounterMetricsType,
		Delta: &delta,
	}

	payload, _ := json.Marshal(metric)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(payload))
		w := httptest.NewRecorder()

		handler.Handle(w, req)
	}
}

// BenchmarkStoreHandler_Gauge measures performance of storing gauge metrics
func BenchmarkStoreHandler_Gauge(b *testing.B) {
	service := newMockMetricsService()
	cfg := &config.Config{}
	handler := NewStoreHandler(service, cfg, nil)

	value := 45.5
	metric := dto.Metrics{
		ID:    "test_gauge",
		MType: dto.GaugeMetricsType,
		Value: &value,
	}

	payload, _ := json.Marshal(metric)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(payload))
		w := httptest.NewRecorder()

		handler.Handle(w, req)
	}
}

// BenchmarkStoreHandler_Parallel measures concurrent metric storage
func BenchmarkStoreHandler_Parallel(b *testing.B) {
	service := newMockMetricsService()
	cfg := &config.Config{}
	handler := NewStoreHandler(service, cfg, nil)

	delta := int64(1)
	metric := dto.Metrics{
		ID:    "test_counter",
		MType: dto.CounterMetricsType,
		Delta: &delta,
	}

	payload, _ := json.Marshal(metric)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(payload))
			w := httptest.NewRecorder()

			handler.Handle(w, req)
		}
	})
}

// BenchmarkStoreAllHandler measures batch metric storage performance
func BenchmarkStoreAllHandler(b *testing.B) {
	service := newMockMetricsService()
	cfg := &config.Config{}
	handler := NewStoreAllHandler(service, cfg, nil)

	sizes := []struct {
		name  string
		count int
	}{
		{"10_metrics", 10},
		{"100_metrics", 100},
		{"1000_metrics", 1000},
	}

	for _, s := range sizes {
		b.Run(s.name, func(b *testing.B) {
			metrics := make([]dto.Metrics, s.count)
			for i := 0; i < s.count; i++ {
				if i%2 == 0 {
					delta := int64(i)
					metrics[i] = dto.Metrics{
						ID:    "counter",
						MType: dto.CounterMetricsType,
						Delta: &delta,
					}
				} else {
					value := float64(i)
					metrics[i] = dto.Metrics{
						ID:    "gauge",
						MType: dto.GaugeMetricsType,
						Value: &value,
					}
				}
			}

			payload, _ := json.Marshal(metrics)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(payload))
				w := httptest.NewRecorder()

				handler.Handle(w, req)
			}
		})
	}
}

// BenchmarkStoreAllHandler_Parallel measures concurrent batch storage
func BenchmarkStoreAllHandler_Parallel(b *testing.B) {
	service := newMockMetricsService()
	cfg := &config.Config{}
	handler := NewStoreAllHandler(service, cfg, nil)

	metrics := make([]dto.Metrics, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			delta := int64(i)
			metrics[i] = dto.Metrics{
				ID:    "counter",
				MType: dto.CounterMetricsType,
				Delta: &delta,
			}
		} else {
			value := float64(i)
			metrics[i] = dto.Metrics{
				ID:    "gauge",
				MType: dto.GaugeMetricsType,
				Value: &value,
			}
		}
	}

	payload, _ := json.Marshal(metrics)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(payload))
			w := httptest.NewRecorder()

			handler.Handle(w, req)
		}
	})
}

// BenchmarkGetHandler measures metric retrieval performance
func BenchmarkGetHandler(b *testing.B) {
	service := newMockMetricsService()
	_ = service.StoreCounter("test_counter", 100)
	_ = service.StoreGauge("test_gauge", 45.5)

	handler := NewGetHandler(service)

	testCases := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "counter",
			payload: []byte(`{"id":"test_counter","type":"counter"}`),
		},
		{
			name:    "gauge",
			payload: []byte(`{"id":"test_gauge","type":"gauge"}`),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(tc.payload))
				w := httptest.NewRecorder()

				handler.Handle(w, req)
			}
		})
	}
}

// BenchmarkGetHandler_Parallel measures concurrent metric retrieval
func BenchmarkGetHandler_Parallel(b *testing.B) {
	service := newMockMetricsService()
	_ = service.StoreCounter("test_counter", 100)

	handler := NewGetHandler(service)
	payload := []byte(`{"id":"test_counter","type":"counter"}`)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(payload))
			w := httptest.NewRecorder()

			handler.Handle(w, req)
		}
	})
}

// BenchmarkEndToEnd simulates a complete store and retrieve cycle
func BenchmarkEndToEnd(b *testing.B) {
	service := newMockMetricsService()
	cfg := &config.Config{}
	storeHandler := NewStoreHandler(service, cfg, nil)
	getHandler := NewGetHandler(service)

	delta := int64(1)
	storeMetric := dto.Metrics{
		ID:    "test_counter",
		MType: dto.CounterMetricsType,
		Delta: &delta,
	}
	storePayload, _ := json.Marshal(storeMetric)

	getMetric := dto.Metrics{
		ID:    "test_counter",
		MType: dto.CounterMetricsType,
	}
	getPayload, _ := json.Marshal(getMetric)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storeReq := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(storePayload))
		storeW := httptest.NewRecorder()
		storeHandler.Handle(storeW, storeReq)

		getReq := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(getPayload))
		getW := httptest.NewRecorder()
		getHandler.Handle(getW, getReq)
	}
}
