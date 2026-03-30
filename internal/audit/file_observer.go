package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// FileObserver добавляет аудит-события в файл в формате JSON (по одному на строку).
type FileObserver struct {
	filePath string
	mu       sync.Mutex
}

// NewFileObserver создаёт FileObserver, записывающий события в указанный файл.
func NewFileObserver(filePath string) *FileObserver {
	return &FileObserver{filePath: filePath}
}

// Notify записывает событие в файл аудита.
func (f *FileObserver) Notify(event AuditEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal audit event to JSON: %w", err)
	}

	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("could not open audit file: %w", err)
	}
	defer file.Close()

	data = append(data, '\n')
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("could not write audit event to file: %w", err)
	}

	return nil
}
