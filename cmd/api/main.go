// @title           Capital Pipefy API
// @version         1.0
// @description     API para gestao de clientes e integracao com Pipefy
// @BasePath        /

package main

import (
	"os"

	"go.uber.org/zap"
	_ "github.com/jalvess021/capital-pipefy/docs"
	"github.com/jalvess021/capital-pipefy/internal/bootstrap"
	"github.com/jalvess021/capital-pipefy/internal/logger"
	"github.com/jalvess021/capital-pipefy/internal/route"
)

func main() {
	log, err := logger.New()
	if err != nil {
		panic(err)
	}

	app, err := bootstrap.NewApp(log)
	if err != nil {
		logger.ApplicationError(log, "failed to bootstrap application", err)
		os.Exit(1)
	}
	defer app.DB.Close()

	logger.ApplicationInfo(log, "application starting",
		zap.String("port", app.Config.Port),
		zap.String("version", "1.0.0"),
	)

	router := route.SetupRouter(app)

	logger.ApplicationInfo(log, "server ready",
		zap.String("port", app.Config.Port),
	)

	if err := router.Run(":" + app.Config.Port); err != nil {
		logger.ApplicationError(log, "failed to start server", err,
			zap.String("port", app.Config.Port),
		)
		os.Exit(1)
	}
}
