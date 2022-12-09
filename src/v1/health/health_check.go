package health

import (
	"encoding/json"
	"time"

	"docker-operator/version"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var upSince = time.Now()

type healthCheck struct {
	Alive     bool   `json:"alive"`
	Since     string `json:"since"`
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Commit    string `json:"commit"`
}

// CheckHandler : A very simple health check.
func CheckHandler(c *fiber.Ctx) error {
	zap.S().Debug("handle health check")
	healthCheck := &healthCheck{
		Alive:     true,
		Since:     time.Since(upSince).String(),
		Version:   version.Version,
		BuildDate: version.BuildDate,
		GoVersion: version.GoVersion,
		Commit:    version.GitCommit,
	}
	responseBytes, err := json.Marshal(healthCheck)
	if err != nil {
		zap.S().Error("Error marshalling healthCheck response: ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err,
		})
	}
	return c.Send(responseBytes)
}
