// Package models описывает модели данных для хранения и передачи метрик.
package models

import "fmt"

const (
	// Counter — тип метрики-счётчика (полное значение).
	Counter = "counter"
	// Gauge — тип метрики-измерения (текущее значение).
	Gauge = "gauge"
)

// Metrics описывает единицу метрики, которая может быть типа Counter или Gauge.
// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

// String возвращает строковое представление метрики для отладки.
func (m *Metrics) String() string {
	return fmt.Sprintf("metrics{ID: %s, MType: %s, Delta: %v, Value: %v, Hash: %s}", m.ID, m.MType, m.Delta, m.Value, m.Hash)
}
