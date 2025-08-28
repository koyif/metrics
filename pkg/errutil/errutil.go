package errutil

import (
	"fmt"
	"github.com/koyif/metrics/pkg/logger"
	"time"
)

type RetriableErrorClassification int

const (
	Retriable RetriableErrorClassification = iota
	NonRetriable
)

type classifier interface {
	Classify(error) RetriableErrorClassification
}

func Retry(classifier classifier, fn func() error) error {
	maxAttempts := 3
	var lastErr error

	for i := 0; i < maxAttempts; i++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		} else {
			if classifier.Classify(lastErr) == NonRetriable {
				return lastErr
			}

			logger.Log.Warn("failed to execute query, retrying")
			time.Sleep(time.Duration(i) * ((time.Second * 2) + 1))
			continue
		}
	}

	return fmt.Errorf("failed to execute query after %d attempts, last error: %w", maxAttempts, lastErr)
}
