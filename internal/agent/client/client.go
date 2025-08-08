package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/pkg/dto"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
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
	metrics := dto.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	err := addValue(&metrics, value)
	if err != nil {
		return err
	}

	return c.sendMetric(metrics)
}

func (c *MetricsClient) sendMetric(metrics dto.Metrics) error {
	requestBody, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	response, err := c.httpClient.Post(
		c.baseURL.String(),
		"application/json",
		bytes.NewReader(requestBody),
	)

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

func addValue(metrics *dto.Metrics, value string) error {
	switch metrics.MType {
	case "counter":
		del, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		(*metrics).Delta = &del
	case "gauge":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		(*metrics).Value = &val
	default:
		return fmt.Errorf("unknown metrics type: %s", metrics.MType)
	}

	return nil
}
