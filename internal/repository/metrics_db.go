package repository

import (
	"database/sql"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
)

type Repository interface {
	Ping() error
	Update(metric *models.Metrics) error
	Get(metricName string) (*models.Metrics, error)
	GetAll() ([]*models.Metrics, error)
}

type MetricsRepository struct {
	db *sql.DB
}

func NewMetricsRepository(db *sql.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func (r *MetricsRepository) DB() *sql.DB {
	return r.db
}

func (r *MetricsRepository) Ping() error {
	return r.DB().Ping()
}

func (r *MetricsRepository) Update(metric *models.Metrics) error {
	gaugeQuery := `
	INSERT INTO metrics (id, metric_type, delta, value)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id)
	DO UPDATE SET
	   value = EXCLUDED.value,
	   updated_at = CURRENT_TIMESTAMP
	`

	counterQuery := `
	INSERT INTO metrics (id, metric_type, delta, value)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id)
	DO UPDATE SET
	   delta = metrics.delta + EXCLUDED.delta,
	   updated_at = CURRENT_TIMESTAMP
	`

	if metric.MType == models.Gauge {
		_, err := r.db.Exec(gaugeQuery,
			metric.ID,
			metric.MType,
			nil,
			metric.Value,
		)
		if err != nil {
			return err
		}
		return nil
	}

	if metric.MType == models.Counter {
		_, err := r.db.Exec(counterQuery,
			metric.ID,
			metric.MType,
			metric.Delta,
			nil,
		)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (r *MetricsRepository) Get(metricName string) (*models.Metrics, error) {
	query := "SELECT id, metric_type, delta, value FROM metrics WHERE id = $1"
	row := r.db.QueryRow(query, metricName)
	var metric models.Metrics
	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

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
		err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			continue
		}
		metrics = append(metrics, &metric)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return metrics, nil
}
