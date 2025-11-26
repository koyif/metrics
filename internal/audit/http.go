package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/pkg/logger"
)

// HTTPAuditor sends audit events to a remote server via HTTP
type HTTPAuditor struct {
	url    string
	client *http.Client
}

// NewHTTPAuditor creates a new HTTP-based audit observer
func NewHTTPAuditor(url string) (*HTTPAuditor, error) {
	return &HTTPAuditor{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

// Notify sends the audit event to the remote server
func (h *HTTPAuditor) Notify(event models.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("failed to marshal audit event", logger.Error(err))
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewBuffer(data))
	if err != nil {
		logger.Log.Error("failed to create audit request", logger.Error(err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		logger.Log.Error("failed to send audit event", logger.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.Log.Warn("audit server returned error status", logger.Int("status", resp.StatusCode))
	}

	return nil
}
