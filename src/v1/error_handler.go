package routes

import (
	"docker-operator/src/common"

	"github.com/gofiber/fiber/v2"
)

// ErrorResp represents the structure of a standard error response.
type ErrorResp struct {
	S      string `json:"s"`
	ErrMsg string `json:"errmsg"`
}

// StandardErrorHandler generates a standardized error message
func StandardErrorHandler(ctx *fiber.Ctx, err error) error {
	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	switch err := err.(type) {
	case common.HttpError:
		return ctx.Status(err.StatusCode()).JSON(ErrorResp{S: "error", ErrMsg: err.Error()})
	default:
		return ctx.Status(200).JSON(ErrorResp{S: "error", ErrMsg: err.Error()})
	}
}
