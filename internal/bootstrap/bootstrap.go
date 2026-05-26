package bootstrap

import (
    "github.com/jalvess021/capital-pipefy/internal/config"
    "github.com/jalvess021/capital-pipefy/internal/database"
    "github.com/jalvess021/capital-pipefy/internal/repository"
    postgresrepo "github.com/jalvess021/capital-pipefy/internal/repository/postgres"
)

type App struct {
    Config     *config.Config
    DB         *database.PostgresDB
    ClientRepo repository.ClientRepository
    EventRepo  repository.EventRepository
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
        Config:     cfg,
        DB:         db,
        ClientRepo: postgresrepo.NewClientRepository(db.GormDB()),
        EventRepo:  postgresrepo.NewEventRepository(db.GormDB()),
    }, nil
}