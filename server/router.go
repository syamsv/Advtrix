package server

import (
	"github.com/gofiber/fiber/v3"

	v1 "github.com/syamsv/Advtrix/api/v1"
)

func mountv1Routes(router fiber.Router) {
	router.Use(AuthMiddleware)

	router.Post("/create", v1.Create)
}
