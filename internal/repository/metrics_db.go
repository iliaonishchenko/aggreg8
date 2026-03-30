// Package repository реализует доступ к данным (PostgreSQL, файловое хранилище).
package repository

import (
	"database/sql"
	"fmt"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"time"
)

// MetricsRepository обеспечивает CRUD-операции с метриками в PostgreSQL с поддержкой ретраев.
type MetricsRepository struct {
	db         *sql.DB
	classifier *PostgresErrorClassifier
}

// NewMetricsRepository создаёт новый репозиторий метрик с указанным подключением к БД.
func NewMetricsRepository(db *sql.DB, classifier *PostgresErrorClassifier) *MetricsRepository {
	return &MetricsRepository{db: db, classifier: classifier}
}

// DB возвращает подключение к базе данных.
func (r *MetricsRepository) DB() *sql.DB {
	return r.db
}

// Ping проверяет доступность базы данных.
func (r *MetricsRepository) Ping() error {
	return r.DB().Ping()
}

// Update вставляет или обновляет метрику в БД (upsert) с ретраями при транзиентных ошибках.
func (r *MetricsRepository) Update(metric *models.Metrics) error {
	return r.executeWithRetry(func() error {
		var query string
		var delta, value interface{}

		switch metric.MType {
		case models.Gauge:
			query = `
				INSERT INTO metrics (id, metric_type, delta, value)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (id)
				DO UPDATE SET
				   value = EXCLUDED.value,
				   updated_at = CURRENT_TIMESTAMP
			`
			delta = nil
			value = metric.Value
		case models.Counter:
			query = `
				INSERT INTO metrics (id, metric_type, delta, value)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (id)
				DO UPDATE SET
				   delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta,
				   updated_at = CURRENT_TIMESTAMP
			`
			delta = metric.Delta
			value = nil
		default:
			return nil
		}

		_, err := r.db.Exec(query, metric.ID, metric.MType, delta, value)
		return err
	})
}

// Get возвращает метрику по имени или nil, если не найдена.
func (r *MetricsRepository) Get(metricName string) (*models.Metrics, error) {
	query := "SELECT id, metric_type, delta, value FROM metrics WHERE id = $1"
	row := r.db.QueryRow(query, metricName)

	var metric models.Metrics
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := row.Scan(&metric.ID, &metric.MType, &delta, &value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if delta.Valid {
		metric.Delta = &delta.Int64
	}
	if value.Valid {
		metric.Value = &value.Float64
	}

	return &metric, nil
}

// GetAll возвращает все метрики из базы данных.
func (r *MetricsRepository) GetAll() ([]*models.Metrics, error) {
	query := "SELECT id, metric_type, delta, value FROM metrics"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var metrics []*models.Metrics
	for rows.Next() {
		var metric models.Metrics
		var delta sql.NullInt64
		var value sql.NullFloat64

		err := rows.Scan(&metric.ID, &metric.MType, &delta, &value)
		if err != nil {
			continue
		}

		if delta.Valid {
			metric.Delta = &delta.Int64
		}
		if value.Valid {
			metric.Value = &value.Float64
		}

		metrics = append(metrics, &metric)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return metrics, nil
}

// BatchUpdate выполняет пакетное обновление метрик в одной транзакции с ретраями.
func (r *MetricsRepository) BatchUpdate(metrics []*models.Metrics) error {
	return r.executeWithRetry(func() error {
		tx, err := r.db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		gaugeStmt, err := tx.Prepare(`
		INSERT INTO metrics (id, metric_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id)
		DO UPDATE SET
		   value = EXCLUDED.value,
		   updated_at = CURRENT_TIMESTAMP
	`)
		if err != nil {
			return err
		}
		defer gaugeStmt.Close()

		counterStmt, err := tx.Prepare(`
		INSERT INTO metrics (id, metric_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id)
		DO UPDATE SET
		   delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta,
		   updated_at = CURRENT_TIMESTAMP
	`)
		if err != nil {
			return err
		}
		defer counterStmt.Close()

		for _, metric := range metrics {
			if metric.MType == models.Gauge {
				_, err := gaugeStmt.Exec(
					metric.ID,
					metric.MType,
					nil,
					metric.Value,
				)
				if err != nil {
					return err
				}
			} else if metric.MType == models.Counter {
				_, err := counterStmt.Exec(
					metric.ID,
					metric.MType,
					metric.Delta,
					nil,
				)
				if err != nil {
					return err
				}
			}
		}

		return tx.Commit()
	})
}

func (r *MetricsRepository) executeWithRetry(operation func() error) error {
	const (
		maxAttempts = 4
		deltaDelay  = 2 * time.Second
	)
	currDelay := 1 * time.Second
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt != 0 {
			time.Sleep(currDelay)
			currDelay += deltaDelay
		}

		err := operation()
		if err == nil {
			return nil
		}

		if !r.classifier.IsRetriable(err) {
			return err
		}
		lastErr = err
	}
	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}
