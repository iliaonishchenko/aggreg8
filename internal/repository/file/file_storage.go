package file

import (
	"encoding/json"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"os"
	"sync"
)

type FileStorage struct {
	mu sync.Mutex
}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

func (fs *FileStorage) SaveToFile(filename string, metrics []*models.Metrics) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Error("error marshaling metrics to JSON", logger.Err(err))
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (fs *FileStorage) ReadFromFile(filename string) ([]*models.Metrics, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Metrics{}, nil
		}
		return nil, err
	}

	var metrics []*models.Metrics
	if err = json.Unmarshal(data, &metrics); err != nil {
		logger.Log.Error("error unmarshaling metrics from JSON", logger.Err(err))
		return nil, err
	}
	return metrics, nil
}
