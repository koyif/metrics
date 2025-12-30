package converter

import (
	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/internal/proto/api/proto"
)

// ProtoToModels converts protobuf Metric messages to internal models.Metrics.
func ProtoToModels(protoMetrics []*proto.Metric) []models.Metrics {
	result := make([]models.Metrics, 0, len(protoMetrics))

	for _, pm := range protoMetrics {
		m := models.Metrics{
			ID: pm.Id,
		}

		switch pm.Type {
		case proto.Metric_COUNTER:
			m.MType = models.Counter
			if pm.Delta != 0 {
				delta := pm.Delta
				m.Delta = &delta
			}
		case proto.Metric_GAUGE:
			m.MType = models.Gauge
			if pm.Value != 0 {
				value := pm.Value
				m.Value = &value
			}
		}

		result = append(result, m)
	}

	return result
}

// ModelsToProto converts internal models.Metrics to protobuf Metric messages.
func ModelsToProto(modelMetrics []models.Metrics) []*proto.Metric {
	result := make([]*proto.Metric, 0, len(modelMetrics))

	for _, mm := range modelMetrics {
		pm := &proto.Metric{
			Id: mm.ID,
		}

		switch mm.MType {
		case models.Counter:
			pm.Type = proto.Metric_COUNTER
			if mm.Delta != nil {
				pm.Delta = *mm.Delta
			}
		case models.Gauge:
			pm.Type = proto.Metric_GAUGE
			if mm.Value != nil {
				pm.Value = *mm.Value
			}
		}

		result = append(result, pm)
	}

	return result
}
