package database

import (
    "fmt"

    gormpostgres "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type PostgresDB struct {
    db *gorm.DB
}

func NewPostgres(url string) (*PostgresDB, error) {
    db, err := gorm.Open(gormpostgres.Open(url), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) GormDB() *gorm.DB { return p.db }

func (p *PostgresDB) Close() error {
    sql, err := p.db.DB()
    if err != nil {
        return err
    }
    return sql.Close()
}