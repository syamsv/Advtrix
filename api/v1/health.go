package v1

import "github.com/gofiber/fiber/v3"

func Health(c fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"status": "READY",
	})
}
