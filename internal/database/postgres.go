package database

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
)

type PostgresDB struct {
    db *sql.DB
}

func NewPostgres(url string) (*PostgresDB, error) {
    db, err := sql.Open("postgres", url)
    if err != nil {
        return nil, fmt.Errorf("failed to open postgres: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping postgres: %w", err)
    }

    return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) Ping() error  { return p.db.Ping() }
func (p *PostgresDB) Close() error { return p.db.Close() }
func (p *PostgresDB) DB() *sql.DB  { return p.db }
