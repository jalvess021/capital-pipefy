package bootstrap

import (
	"github.com/jalvess021/capital-pipefy/internal/config"
	"github.com/jalvess021/capital-pipefy/internal/database"
	"github.com/jalvess021/capital-pipefy/internal/handler"
	pipefyclient "github.com/jalvess021/capital-pipefy/internal/infrastructure/pipefy"
	postgresrepo "github.com/jalvess021/capital-pipefy/internal/repository/postgres"
	"github.com/jalvess021/capital-pipefy/internal/service"
)

type Providers struct {
	ClientHandler  *handler.ClientHandler
	WebhookHandler *handler.WebhookHandler
}

func buildProviders(db *database.PostgresDB, cfg *config.Config) *Providers {
	clientRepo := postgresrepo.NewClientRepository(db.GormDB())
	eventRepo := postgresrepo.NewEventRepository(db.GormDB())

	pipefy := pipefyclient.NewClient(cfg.PipefyAPIURL, cfg.PipefyToken, cfg.PipefyPipeID)

	clientSvc := service.NewClientService(clientRepo, pipefy)
	webhookSvc := service.NewWebhookService(clientRepo, eventRepo)

	return &Providers{
		ClientHandler:  handler.NewClientHandler(clientSvc),
		WebhookHandler: handler.NewWebhookHandler(webhookSvc),
	}
}
