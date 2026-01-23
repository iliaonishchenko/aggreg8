package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestPostgresErrorClassifier_isRetriable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error - not retriable",
			err:  nil,
			want: false,
		},
		{
			name: "generic error - not retriable",
			err:  errors.New("some generic error"),
			want: false,
		},
		{
			name: "08000 connection_exception - retriable",
			err: &pgconn.PgError{
				Code:    "08000",
				Message: "connection_exception",
			},
			want: true,
		},
		{
			name: "08003 connection_does_not_exist - retriable",
			err: &pgconn.PgError{
				Code:    "08003",
				Message: "connection_does_not_exist",
			},
			want: true,
		},
		{
			name: "08006 connection_failure - retriable",
			err: &pgconn.PgError{
				Code:    "08006",
				Message: "connection_failure",
			},
			want: true,
		},
		{
			name: "08001 sqlclient_unable_to_establish_sqlconnection - retriable",
			err: &pgconn.PgError{
				Code:    "08001",
				Message: "sqlclient_unable_to_establish_sqlconnection",
			},
			want: true,
		},
		{
			name: "08P01 protocol_violation - retriable",
			err: &pgconn.PgError{
				Code:    "08P01",
				Message: "protocol_violation",
			},
			want: true,
		},
		{
			name: "23505 unique_violation - not retriable",
			err: &pgconn.PgError{
				Code:    "23505",
				Message: "unique_violation",
			},
			want: false,
		},
		{
			name: "42P01 undefined_table - not retriable",
			err: &pgconn.PgError{
				Code:    "42P01",
				Message: "undefined_table",
			},
			want: false,
		},
		{
			name: "22012 division_by_zero - not retriable",
			err: &pgconn.PgError{
				Code:    "22012",
				Message: "division_by_zero",
			},
			want: false,
		},
		{
			name: "short error code - not retriable",
			err: &pgconn.PgError{
				Code:    "1",
				Message: "short code",
			},
			want: false,
		},
		{
			name: "wrapped postgres connection error - retriable",
			err:  errors.Join(errors.New("outer error"), &pgconn.PgError{Code: "08006"}),
			want: true,
		},
	}

	classifier := NewPostgresErrorClassifier()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.isRetriable(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
