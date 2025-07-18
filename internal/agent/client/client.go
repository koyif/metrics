package client

import (
	"fmt"
	"github.com/koyif/metrics/internal/agent/config"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type MetricsClient struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func New(cfg *config.Config, c *http.Client) (*MetricsClient, error) {
	baseURL, err := url.Parse(fmt.Sprintf("http://%s", cfg.Server.Addr))
	if err != nil {
		return nil, fmt.Errorf("error creating MetricsClient: %w", err)
	}

	baseURL = baseURL.JoinPath("update")

	return &MetricsClient{
		httpClient: c,
		baseURL:    baseURL,
	}, nil
}

func (c *MetricsClient) Send(metricType, metricName, value string) error {
	u := c.baseURL.
		JoinPath(metricType).
		JoinPath(metricName).
		JoinPath(value)

	response, err := c.httpClient.Post(u.String(), "text/plain", nil)
	if err != nil {
		if response != nil && response.Body != nil {
			err := response.Body.Close()
			if err != nil {
				slog.Error(fmt.Sprintf("error closing response body: %v", err))
			}
		}
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error(fmt.Sprintf("error closing response body: %v", err))
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("incorrect response status from Metrics Server: %d", response.StatusCode)
	}

	return nil
}
