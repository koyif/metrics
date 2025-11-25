package metrics_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/handler/metrics"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
	"github.com/koyif/metrics/pkg/dto"
	"github.com/koyif/metrics/pkg/types"
)

// Example_storeCounter demonstrates how to store a counter metric using the JSON API.
// Counter metrics are cumulative values that increase over time.
func Example_storeCounter() {
	// Create test dependencies
	repo := repository.NewMetricsRepository()
	svc := service.NewMetricsService(repo, nil)
	cfg := &config.Config{
		Storage: config.StorageConfig{
			StoreInterval: types.DurationInSeconds(300 * time.Second), // Non-zero to skip immediate persistence
		},
	}
	handler := metrics.NewStoreHandler(svc, cfg, nil)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(handler.Handle))
	defer ts.Close()

	// Prepare request payload
	metric := dto.Metrics{
		ID:    "requests_total",
		MType: "counter",
		Delta: func() *int64 { v := int64(5); return &v }(),
	}

	body, _ := json.Marshal(metric)

	// Send POST request
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// Example_storeGauge demonstrates how to store a gauge metric using the JSON API.
// Gauge metrics represent current state values that can increase or decrease.
func Example_storeGauge() {
	// Create test dependencies
	repo := repository.NewMetricsRepository()
	svc := service.NewMetricsService(repo, nil)
	cfg := &config.Config{
		Storage: config.StorageConfig{
			StoreInterval: types.DurationInSeconds(300 * time.Second), // Non-zero to skip immediate persistence
		},
	}
	handler := metrics.NewStoreHandler(svc, cfg, nil)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(handler.Handle))
	defer ts.Close()

	// Prepare request payload
	metric := dto.Metrics{
		ID:    "cpu_usage",
		MType: "gauge",
		Value: func() *float64 { v := 75.5; return &v }(),
	}

	body, _ := json.Marshal(metric)

	// Send POST request
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// Example_batchStore demonstrates how to store multiple metrics in a single batch request.
// This is more efficient than individual requests when updating many metrics at once.
func Example_batchStore() {
	// Create test dependencies
	repo := repository.NewMetricsRepository()
	svc := service.NewMetricsService(repo, nil)
	cfg := &config.Config{
		Storage: config.StorageConfig{
			StoreInterval: types.DurationInSeconds(300 * time.Second), // Non-zero to skip immediate persistence
		},
	}
	handler := metrics.NewStoreAllHandler(svc, cfg, nil)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(handler.Handle))
	defer ts.Close()

	// Prepare batch request payload
	metricsData := []dto.Metrics{
		{
			ID:    "requests_total",
			MType: "counter",
			Delta: func() *int64 { v := int64(10); return &v }(),
		},
		{
			ID:    "memory_usage",
			MType: "gauge",
			Value: func() *float64 { v := 512.0; return &v }(),
		},
		{
			ID:    "errors_total",
			MType: "counter",
			Delta: func() *int64 { v := int64(2); return &v }(),
		},
	}

	body, _ := json.Marshal(metricsData)

	// Send POST request
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// Example_getCounter demonstrates how to retrieve a counter metric value.
func Example_getCounter() {
	// Create test dependencies
	repo := repository.NewMetricsRepository()
	svc := service.NewMetricsService(repo, nil)

	// Pre-populate with test data
	_ = svc.StoreCounter("requests_total", 42)

	handler := metrics.NewGetHandler(svc)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(handler.Handle))
	defer ts.Close()

	// Prepare request payload
	metric := dto.Metrics{
		ID:    "requests_total",
		MType: "counter",
	}

	body, _ := json.Marshal(metric)

	// Send POST request
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read response
	var result dto.Metrics
	respBody, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(respBody, &result)

	fmt.Printf("Metric: %s, Type: %s, Value: %d\n", result.ID, result.MType, *result.Delta)
	// Output: Metric: requests_total, Type: counter, Value: 42
}

// Example_getGauge demonstrates how to retrieve a gauge metric value.
func Example_getGauge() {
	// Create test dependencies
	repo := repository.NewMetricsRepository()
	svc := service.NewMetricsService(repo, nil)

	// Pre-populate with test data
	_ = svc.StoreGauge("cpu_usage", 85.3)

	handler := metrics.NewGetHandler(svc)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(handler.Handle))
	defer ts.Close()

	// Prepare request payload
	metric := dto.Metrics{
		ID:    "cpu_usage",
		MType: "gauge",
	}

	body, _ := json.Marshal(metric)

	// Send POST request
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Read response
	var result dto.Metrics
	respBody, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(respBody, &result)

	fmt.Printf("Metric: %s, Type: %s, Value: %.1f\n", result.ID, result.MType, *result.Value)
	// Output: Metric: cpu_usage, Type: gauge, Value: 85.3
}

// Example_counterAccumulation demonstrates how counter metrics accumulate over multiple updates.
func Example_counterAccumulation() {
	// Create test dependencies
	repo := repository.NewMetricsRepository()
	svc := service.NewMetricsService(repo, nil)
	cfg := &config.Config{
		Storage: config.StorageConfig{
			StoreInterval: types.DurationInSeconds(300 * time.Second), // Non-zero to skip immediate persistence
		},
	}
	storeHandler := metrics.NewStoreHandler(svc, cfg, nil)
	getHandler := metrics.NewGetHandler(svc)

	storeServer := httptest.NewServer(http.HandlerFunc(storeHandler.Handle))
	defer storeServer.Close()

	// Send first update
	metric1 := dto.Metrics{
		ID:    "total_requests",
		MType: "counter",
		Delta: func() *int64 { v := int64(10); return &v }(),
	}
	body1, _ := json.Marshal(metric1)
	_, _ = http.Post(storeServer.URL, "application/json", bytes.NewReader(body1))

	// Send second update
	metric2 := dto.Metrics{
		ID:    "total_requests",
		MType: "counter",
		Delta: func() *int64 { v := int64(25); return &v }(),
	}
	body2, _ := json.Marshal(metric2)
	_, _ = http.Post(storeServer.URL, "application/json", bytes.NewReader(body2))

	// Retrieve accumulated value
	getServer := httptest.NewServer(http.HandlerFunc(getHandler.Handle))
	defer getServer.Close()

	getMetric := dto.Metrics{
		ID:    "total_requests",
		MType: "counter",
	}
	getBody, _ := json.Marshal(getMetric)
	resp, _ := http.Post(getServer.URL, "application/json", bytes.NewReader(getBody))
	defer resp.Body.Close()

	var result dto.Metrics
	respBody, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(respBody, &result)

	fmt.Printf("Total accumulated value: %d\n", *result.Delta)
	// Output: Total accumulated value: 35
}