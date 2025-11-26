package audit

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/pkg/logger"
)

// FileAuditor writes audit events to a file
type FileAuditor struct {
	filePath string
	mu       sync.Mutex
}

// NewFileAuditor creates a new file-based audit observer
func NewFileAuditor(filePath string) (*FileAuditor, error) {
	return &FileAuditor{
		filePath: filePath,
	}, nil
}

// Notify writes the audit event to the file
func (f *FileAuditor) Notify(event models.AuditEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Log.Error("failed to open audit file", logger.Error(err))
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.Log.Error("error closing audit file", logger.Error(err))
		}
	}(file)

	data, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("failed to marshal audit event", logger.Error(err))
		return err
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		logger.Log.Error("failed to write audit event to file", logger.Error(err))
		return err
	}

	return nil
}
