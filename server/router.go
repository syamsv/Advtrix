package server

import (
	"github.com/gofiber/fiber/v3"

	v1 "github.com/syamsv/Advtrix/api/v1"
)

func mountv1Routes(router fiber.Router) {
	router.Use(AuthMiddleware)
	router.Use(NTSHealthMiddleware)

	router.Post("/create", v1.Create)
	router.Post("/getcode", v1.GetCode)
	router.Post("/validate", v1.Validate)
}
