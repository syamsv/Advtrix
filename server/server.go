package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"go.uber.org/zap"

	v1 "github.com/syamsv/Advtrix/api/v1"
	"github.com/syamsv/Advtrix/config"
)

var app *fiber.App

func Run(address string) {
	app = fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: false,
		ServerHeader:  config.APP_NAME,
		AppName:       config.APP_NAME,
	})

	app.Use(RecoverMiddleware)
	app.Use(RequestIDMiddleware)
	app.Use(LoggingMiddleware)
	app.Use(cors.New())

	app.Get("/health", v1.Health)

	v1Group := app.Group("/v1")
	mountv1Routes(v1Group)

	if err := app.Listen(address); err != nil {
		zap.L().Error("server error", zap.Error(err))
	}
}

func Shutdown() {
	if app != nil {
		if err := app.Shutdown(); err != nil {
			zap.L().Error("error during shutdown", zap.Error(err))
		}
	}
}
