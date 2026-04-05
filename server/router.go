package server

import "github.com/gofiber/fiber/v3"

func mountv1Routes(router fiber.Router) {
	router.Use(AuthMiddleware)

	// Add authenticated v1 routes here
}
