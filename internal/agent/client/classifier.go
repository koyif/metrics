package client

import (
	"errors"
	"github.com/koyif/metrics/pkg/errutil"
	"syscall"
)

type HttpErrorClassifier struct{}

func NewHttpErrorClassifier() *HttpErrorClassifier {
	return &HttpErrorClassifier{}
}

func (c *HttpErrorClassifier) Classify(err error) errutil.RetriableErrorClassification {
	if err == nil {
		return errutil.NonRetriable
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return errutil.Retriable
	}

	return errutil.NonRetriable
}
