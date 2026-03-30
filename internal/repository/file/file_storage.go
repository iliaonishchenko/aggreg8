// Package file реализует файловое хранилище метрик в формате JSON.
package file

import (
	"encoding/json"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"os"
	"sync"
)

// FileStorage обеспечивает потокобезопасное сохранение и чтение метрик из JSON-файла.
type FileStorage struct {
	mu sync.Mutex
}

// NewFileStorage создаёт новый FileStorage.
func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

// SaveToFile сериализует метрики в JSON и записывает в указанный файл.
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

// ReadFromFile читает метрики из JSON-файла. Если файл не существует, возвращает пустой срез.
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
