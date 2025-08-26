package repository

import (
	"encoding/json"
	"github.com/koyif/metrics/pkg/logger"
	"os"

	models "github.com/koyif/metrics/internal/models"
)

type FileRepository struct {
	filePath string
}

func NewFileRepository(filePath string) *FileRepository {
	return &FileRepository{
		filePath: filePath,
	}
}

func (r *FileRepository) Save(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	file, err := os.Create(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	json.NewEncoder(file).Encode(metrics)
	return nil
}

func (r *FileRepository) Load() ([]models.Metrics, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Log.Error("error closing file", logger.Error(err))
		}
	}(file)

	var metrics []models.Metrics
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}
