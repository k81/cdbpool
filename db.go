package cdbpool

import (
	"context"
	"database/sql"
)

type DB struct {
	*sql.DB
	ctx context.Context
}

func (db *DB) GetContext() context.Context {
	return db.ctx
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(db.ctx, query, args...)
}

func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(db.ctx, query, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(db.ctx, query, args...)
}

func (db *DB) Begin() (*sql.Tx, error) {
	return db.DB.BeginTx(db.ctx, nil)
}

func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	return db.DB.PrepareContext(db.ctx, query)
}

func (db *DB) Ping() error {
	return db.DB.PingContext(db.ctx)
}
