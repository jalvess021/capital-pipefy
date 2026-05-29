package database

import (
	"fmt"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/jalvess021/capital-pipefy/internal/config"
)

type PostgresDB struct {
	db *gorm.DB
}

func NewPostgres(cfg config.DatabaseConfig) (*PostgresDB, error) {
	db, err := gorm.Open(gormpostgres.Open(cfg.URL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

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
