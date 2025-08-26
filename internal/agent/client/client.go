package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/koyif/metrics/pkg/logger"
	"io"
	"net/http"
	"net/url"

	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/models"
)

type MetricsClient struct {
	httpClient *http.Client
	baseURL    *url.URL
}

const errClosingResponseBody = "error closing response body"

func New(cfg *config.Config, c *http.Client) (*MetricsClient, error) {
	baseURL, err := url.Parse(fmt.Sprintf("http://%s", cfg.Server.Addr))
	if err != nil {
		return nil, fmt.Errorf("error creating MetricsClient: %w", err)
	}

	return &MetricsClient{
		httpClient: c,
		baseURL:    baseURL,
	}, nil
}

func (c *MetricsClient) SendMetric(metric models.Metrics) error {
	requestBody, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	updateURL := c.baseURL.JoinPath("update")

	response, err := c.httpClient.Post(
		updateURL.String(),
		"application/json",
		bytes.NewReader(requestBody),
	)

	if err != nil {
		if response != nil && response.Body != nil {
			err := response.Body.Close()
			if err != nil {
				logger.Log.Error(errClosingResponseBody, logger.Error(err))
			}
		}
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Log.Error(errClosingResponseBody, logger.Error(err))
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("incorrect response status from Metrics Server: %d", response.StatusCode)
	}

	return nil
}

func (c *MetricsClient) SendMetrics(metrics []models.Metrics) error {
	requestBody, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	updatesURL := c.baseURL.JoinPath("updates/")

	response, err := c.httpClient.Post(
		updatesURL.String(),
		"application/json",
		bytes.NewReader(requestBody),
	)

	if err != nil {
		if response != nil && response.Body != nil {
			err := response.Body.Close()
			if err != nil {
				logger.Log.Error(errClosingResponseBody, logger.Error(err))
			}
		}
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Log.Error(errClosingResponseBody, logger.Error(err))
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("incorrect response status from Metrics Server: %d", response.StatusCode)
	}

	return nil
}
