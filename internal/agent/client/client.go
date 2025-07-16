package client

import (
	"fmt"
	"github.com/koyif/metrics/internal/agent/config"
	"net/http"
	"net/url"
)

type MetricsClient struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func New(cfg *config.Config, c *http.Client) (*MetricsClient, error) {
	const op = "MetricsClient.New"
	baseURL, err := url.Parse(fmt.Sprintf("http://%s", cfg.Server.Addr))
	if err != nil {
		return nil, fmt.Errorf("%s: error creating MetricsClient: %w", op, err)
	}

	baseURL = baseURL.JoinPath("update")

	return &MetricsClient{
		httpClient: c,
		baseURL:    baseURL,
	}, nil
}

func (c *MetricsClient) Send(metricType, metricName, value string) error {
	const op = "MetricsClient.Send"
	u := c.baseURL.
		JoinPath(metricType).
		JoinPath(metricName).
		JoinPath(value)

	response, err := c.httpClient.Post(u.String(), "text/plain", nil)
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: incorrect response status: %d", op, response.StatusCode)
	}

	return nil
}
