package client

import (
	"fmt"
	"github.com/koyif/metrics/internal/agent/config"
	"io"
	"net/http"
	"net/url"
)

type MetricsClient struct {
	httpClient *http.Client
	baseUrl    *url.URL
}

func New(cfg *config.Config, c *http.Client) (*MetricsClient, error) {
	const op = "MetricsClient.New"
	baseUrl, err := url.Parse(cfg.Server.Addr)
	if err != nil {
		return nil, fmt.Errorf("%s: error creating MetricsClient: %w", op, err)
	}

	baseUrl = baseUrl.JoinPath("update")

	return &MetricsClient{
		httpClient: c,
		baseUrl:    baseUrl,
	}, nil
}

func (c *MetricsClient) Send(metricType, metricName, value string) error {
	const op = "MetricsClient.Send"
	u := c.baseUrl.
		JoinPath(metricType).
		JoinPath(metricName).
		JoinPath(value)

	response, err := c.httpClient.Post(u.String(), "text/plain", nil)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: incorrect response status: %d", op, response.StatusCode)
	}

	return nil
}
