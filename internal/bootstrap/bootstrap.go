package bootstrap

import (
	"github.com/jalvess021/capital-pipefy/internal/config"
	"github.com/jalvess021/capital-pipefy/internal/database"
)

func NewApp() (*config.Config, database.Database, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	return cfg, db, nil
}