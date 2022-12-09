package routes

import (
	docker2 "docker-operator/src/docker"
	"docker-operator/src/v1/docker"
	"docker-operator/src/v1/health"

	"github.com/gofiber/fiber/v2"
)

func AddRoutes(app *fiber.App, dockerService docker2.ServiceInterface) {
	v1 := app.Group("/api")
	// Health
	v1.Get("/status", health.CheckHandler)
	// Run container
	v1.Get("/exec/:image_name/:tag", docker.RunContainerGet(dockerService))
	v1.Post("/exec/:image_name/:tag", docker.RunContainerPost(dockerService))
}
