package client

import (
	"errors"
	"github.com/koyif/metrics/pkg/errutil"
	"syscall"
)

type HTTPErrorClassifier struct{}

func NewHTTPErrorClassifier() *HTTPErrorClassifier {
	return &HTTPErrorClassifier{}
}

func (c *HTTPErrorClassifier) Classify(err error) errutil.RetriableErrorClassification {
	if err == nil {
		return errutil.NonRetriable
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return errutil.Retriable
	}

	return errutil.NonRetriable
}
