package repository

import "database/sql"

type Database struct {
	db *sql.DB
}

func NewDatabase(db *sql.DB) *Database {
	return &Database{db: db}
}

func (d *Database) DB() *sql.DB {
	return d.db
}

func (d *Database) Ping() error {
	return d.DB().Ping()
}
