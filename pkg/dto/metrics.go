package dto

const (
	// CounterMetricsType is the type identifier for counter metrics.
	CounterMetricsType = "counter"
	// GaugeMetricsType is the type identifier for gauge metrics.
	GaugeMetricsType = "gauge"
)

// Metrics is the data transfer object for metrics API requests and responses.
// It is used for JSON encoding/decoding in HTTP handlers.
//
// This DTO is similar to models.Metrics but excludes the Hash field,
// which is handled separately by middleware.
type Metrics struct {
	// ID is the unique name/identifier of the metric.
	ID string `json:"id"`
	// MType specifies the metric type: "counter" or "gauge".
	MType string `json:"type"`
	// Delta holds the counter value increment. Non-nil only for counter metrics.
	Delta *int64 `json:"delta,omitempty"`
	// Value holds the gauge value. Non-nil only for gauge metrics.
	Value *float64 `json:"value,omitempty"`
}
