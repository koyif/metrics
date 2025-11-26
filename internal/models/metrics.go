package models

const (
	// Counter represents the counter metric type.
	// Counters are cumulative values that only increase over time.
	Counter = "counter"
	// Gauge represents the gauge metric type.
	// Gauges are values that can increase or decrease over time.
	Gauge = "gauge"
)

// Metrics represents a single metric with its metadata and value.
// It supports two types of metrics: counters (cumulative values) and gauges (current values).
//
// The structure uses a flat model without hierarchical nesting for simplicity.
// Delta and Value are declared as pointers to distinguish between a zero value
// and an unset value, allowing proper JSON encoding/decoding with omitempty.
//
// Hash field is used for HMAC validation when a secret key is configured.
type Metrics struct {
	// ID is the unique name/identifier of the metric.
	ID string `json:"id"`
	// MType specifies the metric type: "counter" or "gauge".
	MType string `json:"type"`
	// Delta holds the counter value (cumulative increment). Used only for counter metrics.
	Delta *int64 `json:"delta,omitempty"`
	// Value holds the gauge value (current state). Used only for gauge metrics.
	Value *float64 `json:"value,omitempty"`
	// Hash is the HMAC-SHA256 signature for request validation.
	Hash string `json:"hash,omitempty"`
}
