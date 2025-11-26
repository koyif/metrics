package repository

import (
	"testing"

	"github.com/koyif/metrics/internal/models"
)

// BenchmarkStoreCounter measures the performance of storing counter metrics
func BenchmarkStoreCounter(b *testing.B) {
	repo := NewMetricsRepository()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = repo.StoreCounter("test_counter", int64(i))
	}
}

// BenchmarkStoreCounterParallel measures concurrent counter storage performance
func BenchmarkStoreCounterParallel(b *testing.B) {
	repo := NewMetricsRepository()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = repo.StoreCounter("test_counter", int64(i))
			i++
		}
	})
}

// BenchmarkStoreGauge measures the performance of storing gauge metrics
func BenchmarkStoreGauge(b *testing.B) {
	repo := NewMetricsRepository()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = repo.StoreGauge("test_gauge", float64(i))
	}
}

// BenchmarkStoreGaugeParallel measures concurrent gauge storage performance
func BenchmarkStoreGaugeParallel(b *testing.B) {
	repo := NewMetricsRepository()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = repo.StoreGauge("test_gauge", float64(i))
			i++
		}
	})
}

// BenchmarkCounter measures the performance of reading counter values
func BenchmarkCounter(b *testing.B) {
	repo := NewMetricsRepository()
	_ = repo.StoreCounter("test_counter", 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = repo.Counter("test_counter")
	}
}

// BenchmarkCounterParallel measures concurrent counter reading performance
func BenchmarkCounterParallel(b *testing.B) {
	repo := NewMetricsRepository()
	_ = repo.StoreCounter("test_counter", 100)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = repo.Counter("test_counter")
		}
	})
}

// BenchmarkGauge measures the performance of reading gauge values
func BenchmarkGauge(b *testing.B) {
	repo := NewMetricsRepository()
	_ = repo.StoreGauge("test_gauge", 100.5)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = repo.Gauge("test_gauge")
	}
}

// BenchmarkGaugeParallel measures concurrent gauge reading performance
func BenchmarkGaugeParallel(b *testing.B) {
	repo := NewMetricsRepository()
	_ = repo.StoreGauge("test_gauge", 100.5)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = repo.Gauge("test_gauge")
		}
	})
}

// BenchmarkAllCounters measures the performance of retrieving all counters
func BenchmarkAllCounters(b *testing.B) {
	repo := NewMetricsRepository()
	for i := 0; i < 100; i++ {
		_ = repo.StoreCounter("counter_"+string(rune(i)), int64(i))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = repo.AllCounters()
	}
}

// BenchmarkAllGauges measures the performance of retrieving all gauges
func BenchmarkAllGauges(b *testing.B) {
	repo := NewMetricsRepository()
	for i := 0; i < 100; i++ {
		_ = repo.StoreGauge("gauge_"+string(rune(i)), float64(i))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = repo.AllGauges()
	}
}

// BenchmarkStoreAll measures batch metric storage performance
func BenchmarkStoreAll(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"10_metrics", 10},
		{"100_metrics", 100},
		{"1000_metrics", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			metrics := make([]models.Metrics, bm.count)
			for i := 0; i < bm.count; i++ {
				if i%2 == 0 {
					delta := int64(i)
					metrics[i] = models.Metrics{
						ID:    "counter",
						MType: models.Counter,
						Delta: &delta,
					}
				} else {
					value := float64(i)
					metrics[i] = models.Metrics{
						ID:    "gauge",
						MType: models.Gauge,
						Value: &value,
					}
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				repo := NewMetricsRepository()
				_ = repo.StoreAll(metrics)
			}
		})
	}
}

// BenchmarkStoreAllParallel measures concurrent batch storage performance
func BenchmarkStoreAllParallel(b *testing.B) {
	metrics := make([]models.Metrics, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			delta := int64(i)
			metrics[i] = models.Metrics{
				ID:    "counter",
				MType: models.Counter,
				Delta: &delta,
			}
		} else {
			value := float64(i)
			metrics[i] = models.Metrics{
				ID:    "gauge",
				MType: models.Gauge,
				Value: &value,
			}
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			repo := NewMetricsRepository()
			_ = repo.StoreAll(metrics)
		}
	})
}

// BenchmarkMixedOperations simulates realistic mixed workload
func BenchmarkMixedOperations(b *testing.B) {
	repo := NewMetricsRepository()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = repo.StoreCounter("request_count", 1)
		_ = repo.StoreGauge("cpu_usage", 45.5)
		_, _ = repo.Counter("request_count")
		_, _ = repo.Gauge("cpu_usage")
	}
}
