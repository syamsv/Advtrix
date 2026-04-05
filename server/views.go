package server

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func errorResponse(c fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

func successResponse(c fiber.Ctx, data any) error {
	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func createdResponse(c fiber.Ctx, data any) error {
	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func invalidParams(c fiber.Ctx) error {
	return errorResponse(c, 400, "Invalid parameters: The request was invalid and was unable to be processed")
}

func internalServerError(c fiber.Ctx, err error) error {
	zap.L().Error("internal server error", zap.Error(err), zap.Any("request_id", c.Locals("request_id")))
	return errorResponse(c, 500, "Internal Server Error: Something went wrong on the server side.")
}

func recordNotFound(c fiber.Ctx) error {
	return errorResponse(c, 404, "Not Found: The requested resource was not found.")
}

func unauthorized(c fiber.Ctx) error {
	return errorResponse(c, 401, "Unauthorized: You don't have permission to access this resource.")
}

func forbidden(c fiber.Ctx) error {
	return errorResponse(c, 403, "Forbidden: You don't have permission to access this resource.")
}

func badRequest(c fiber.Ctx) error {
	return errorResponse(c, 400, "Bad request: The request was invalid and was unable to be processed")
}

func conflict(c fiber.Ctx) error {
	return errorResponse(c, 409, "Conflict: The request could not be completed due to a conflict with the current state of the target resource.")
}
