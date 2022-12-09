package main

import (
	"docker-operator/config"
	"docker-operator/log"
	"docker-operator/src/docker"
	routes "docker-operator/src/v1"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func main() {
	loggerMgr := log.InitZapLog()
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync()
	logger := loggerMgr.Sugar()

	dockerClient, err := docker.NewClient()
	if err != nil {
		logger.Fatalf(err.Error())
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: routes.StandardErrorHandler,
	})
	routes.AddRoutes(app, docker.NewService(dockerClient))

	err = app.Listen(config.DefaultConfig.GetString("API_PORT"))
	if err != nil {
		logger.Fatalf(err.Error())
	}
}
