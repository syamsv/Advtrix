package server

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"go.uber.org/zap"

	v1 "github.com/syamsv/go-template/api/v1"
	"github.com/syamsv/go-template/config"
)

var app *fiber.App

func Run(address string) {
	app = fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: false,
		ServerHeader:  config.APP_NAME,
		AppName:       fmt.Sprintf("%s %s", config.APP_NAME, config.APP_VERSION),
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
