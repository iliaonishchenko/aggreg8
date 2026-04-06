package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// FileObserver добавляет аудит-события в файл в формате JSON (по одному на строку).
type FileObserver struct {
	file *os.File
	mu   sync.Mutex
}

// NewFileObserver создаёт FileObserver, записывающий события в указанный файл.
func NewFileObserver(filePath string) (*FileObserver, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not open audit file: %w", err)
	}
	return &FileObserver{file: file}, nil
}

// Close закрывает файл аудита.
func (f *FileObserver) Close() error {
	return f.file.Close()
}

// Notify записывает событие в файл аудита.
func (f *FileObserver) Notify(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal audit event to JSON: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	data = append(data, '\n')
	_, err = f.file.Write(data)
	if err != nil {
		return fmt.Errorf("could not write audit event to file: %w", err)
	}

	return nil
}
