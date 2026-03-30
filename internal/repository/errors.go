package repository

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresErrorClassifier определяет, является ли ошибка PostgreSQL ретраебл.
type PostgresErrorClassifier struct{}

// NewPostgresErrorClassifier создаёт новый PostgresErrorClassifier.
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// IsRetriable возвращает true, если ошибка относится к классу "08" (Connection Exception) в PostgreSQL.
func (c *PostgresErrorClassifier) IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08"
	}

	return false
}
