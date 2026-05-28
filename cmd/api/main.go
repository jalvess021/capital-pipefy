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

	app, err := bootstrap.NewApp()
	if err != nil {
		logger.ApplicationError(log, "Failed to bootstrap application", err)
		os.Exit(1)
	}
	defer app.DB.Close()

	if os.Getenv("PORT") == "" {
		logger.ApplicationWarn(log, "PORT environment variable not set, using default",
			zap.String("port", app.Config.Port),
		)
	}

	logger.ApplicationInfo(log, "Application starting", zap.String("port", app.Config.Port))

	router := route.SetupRouter(app)

	logger.ApplicationInfo(log, "Server ready", zap.String("port", app.Config.Port))

	if err := router.Run(":" + app.Config.Port); err != nil {
		logger.ApplicationError(log, "Failed to start server", err,
			zap.String("port", app.Config.Port),
		)
		os.Exit(1)
	}
}
