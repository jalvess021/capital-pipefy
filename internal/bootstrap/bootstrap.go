package bootstrap

import (
	"github.com/jalvess021/capital-pipefy/internal/config"
	"github.com/jalvess021/capital-pipefy/internal/database"
)

type App struct {
	Config    *config.Config
	DB        *database.PostgresDB
	Providers *Providers
}

func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return &App{
		Config:    cfg,
		DB:        db,
		Providers: buildProviders(db, cfg),
	}, nil
}
