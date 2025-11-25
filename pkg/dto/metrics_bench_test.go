package dto

import (
	"bytes"
	"encoding/json"
	"testing"
)

// BenchmarkMetricsEncode measures JSON encoding performance
func BenchmarkMetricsEncode(b *testing.B) {
	delta := int64(100)
	value := 45.5

	testCases := []struct {
		name    string
		metrics Metrics
	}{
		{
			name: "counter",
			metrics: Metrics{
				ID:    "test_counter",
				MType: CounterMetricsType,
				Delta: &delta,
			},
		},
		{
			name: "gauge",
			metrics: Metrics{
				ID:    "test_gauge",
				MType: GaugeMetricsType,
				Value: &value,
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = json.Marshal(tc.metrics)
			}
		})
	}
}

// BenchmarkMetricsDecode measures JSON decoding performance
func BenchmarkMetricsDecode(b *testing.B) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "counter",
			data: []byte(`{"id":"test_counter","type":"counter","delta":100}`),
		},
		{
			name: "gauge",
			data: []byte(`{"id":"test_gauge","type":"gauge","value":45.5}`),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var m Metrics
				_ = json.Unmarshal(tc.data, &m)
			}
		})
	}
}

// BenchmarkMetricsArrayEncode measures JSON encoding of metric arrays
func BenchmarkMetricsArrayEncode(b *testing.B) {
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
			metrics := make([]Metrics, s.count)
			for i := 0; i < s.count; i++ {
				if i%2 == 0 {
					delta := int64(i)
					metrics[i] = Metrics{
						ID:    "counter",
						MType: CounterMetricsType,
						Delta: &delta,
					}
				} else {
					value := float64(i)
					metrics[i] = Metrics{
						ID:    "gauge",
						MType: GaugeMetricsType,
						Value: &value,
					}
				}
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = json.Marshal(metrics)
			}
		})
	}
}

// BenchmarkMetricsArrayDecode measures JSON decoding of metric arrays
func BenchmarkMetricsArrayDecode(b *testing.B) {
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
			metrics := make([]Metrics, s.count)
			for i := 0; i < s.count; i++ {
				if i%2 == 0 {
					delta := int64(i)
					metrics[i] = Metrics{
						ID:    "counter",
						MType: CounterMetricsType,
						Delta: &delta,
					}
				} else {
					value := float64(i)
					metrics[i] = Metrics{
						ID:    "gauge",
						MType: GaugeMetricsType,
						Value: &value,
					}
				}
			}

			data, _ := json.Marshal(metrics)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var m []Metrics
				_ = json.Unmarshal(data, &m)
			}
		})
	}
}

// BenchmarkMetricsDecoder measures performance of json.Decoder (used in HTTP handlers)
func BenchmarkMetricsDecoder(b *testing.B) {
	testCases := []struct {
		name string
		data string
	}{
		{
			name: "counter",
			data: `{"id":"test_counter","type":"counter","delta":100}`,
		},
		{
			name: "gauge",
			data: `{"id":"test_gauge","type":"gauge","value":45.5}`,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var m Metrics
				reader := bytes.NewReader([]byte(tc.data))
				_ = json.NewDecoder(reader).Decode(&m)
			}
		})
	}
}

// BenchmarkMetricsEncoder measures performance of json.Encoder (used in HTTP handlers)
func BenchmarkMetricsEncoder(b *testing.B) {
	delta := int64(100)
	value := 45.5

	testCases := []struct {
		name    string
		metrics Metrics
	}{
		{
			name: "counter",
			metrics: Metrics{
				ID:    "test_counter",
				MType: CounterMetricsType,
				Delta: &delta,
			},
		},
		{
			name: "gauge",
			metrics: Metrics{
				ID:    "test_gauge",
				MType: GaugeMetricsType,
				Value: &value,
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				_ = json.NewEncoder(&buf).Encode(tc.metrics)
			}
		})
	}
}