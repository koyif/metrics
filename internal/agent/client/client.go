package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/pkg/crypto"
	"github.com/koyif/metrics/pkg/errutil"
	"github.com/koyif/metrics/pkg/logger"
)

type MetricsClient struct {
	httpClient *http.Client
	baseURL    *url.URL
	cfg        *config.Config
	publicKey  *rsa.PublicKey
}

const errClosingResponseBody = "error closing response body"

func New(cfg *config.Config, c *http.Client) (*MetricsClient, error) {
	baseURL, err := url.Parse(fmt.Sprintf("http://%s", cfg.Server.Addr))
	if err != nil {
		return nil, fmt.Errorf("error creating MetricsClient: %w", err)
	}

	client := &MetricsClient{
		httpClient: c,
		baseURL:    baseURL,
		cfg:        cfg,
	}

	if cfg.CryptoKey != "" {
		publicKey, err := crypto.LoadPublicKey(cfg.CryptoKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}
		client.publicKey = publicKey
		logger.Log.Info("public key loaded successfully for encryption")
	}

	return client, nil
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

	return c.retry(updatesURL, requestBody)

}

func (c *MetricsClient) retry(updatesURL *url.URL, requestBody []byte) error {
	maxAttempts := 3
	var lastErr error
	var response *http.Response
	var classifier = NewHTTPErrorClassifier()

	dataToSend := requestBody
	if c.publicKey != nil {
		encryptedData, err := crypto.EncryptData(c.publicKey, requestBody)
		if err != nil {
			return fmt.Errorf("failed to encrypt request body: %w", err)
		}
		dataToSend = encryptedData
	}

	req, err := http.NewRequest(http.MethodPost, updatesURL.String(), bytes.NewReader(dataToSend))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	if c.publicKey != nil {
		req.Header.Set("Content-Type", "application/octet-stream")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.cfg.HashKey != "" {
		h := hmac.New(sha256.New, []byte(c.cfg.HashKey))
		_, err = h.Write(requestBody)
		if err != nil {
			return fmt.Errorf("error creating HMAC: %w", err)
		}

		req.Header.Set("HashSHA256", fmt.Sprintf("%x", h.Sum(nil)))
	}

	for i := range maxAttempts {
		response, lastErr = c.httpClient.Do(req)
		if lastErr == nil {
			if response.StatusCode >= 500 {
				logger.Log.Warn("failed to execute query, retrying")
				time.Sleep(time.Duration(i) * ((time.Second * 2) + 1))
				continue
			}

			err := response.Body.Close()
			if err != nil {
				logger.Log.Error(errClosingResponseBody, logger.Error(err))
			}

			return nil
		} else {
			if classifier.Classify(lastErr) == errutil.NonRetriable {
				return fmt.Errorf("failed to execute query: %w", lastErr)
			}

			logger.Log.Warn("failed to execute query, retrying")
			time.Sleep(time.Duration(i) * ((time.Second * 2) + 1))
			continue
		}
	}

	return fmt.Errorf("failed to execute query after %d attempts, last error: %w", maxAttempts, lastErr)
}
