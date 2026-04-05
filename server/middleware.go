package server

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/syamsv/go-template/config"
)

func AuthMiddleware(c fiber.Ctx) error {
	auth := c.Get("Authorization")
	token, found := strings.CutPrefix(auth, "Bearer ")
	if !found || token != config.INTERNAL_AUTH_PARAMATER {
		return unauthorized(c)
	}
	return c.Next()
}

func RequestIDMiddleware(c fiber.Ctx) error {
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	c.Set("X-Request-ID", requestID)
	c.Locals("request_id", requestID)
	return c.Next()
}

func LoggingMiddleware(c fiber.Ctx) error {
	start := time.Now()

	err := c.Next()

	zap.L().Info("request",
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.Int("status", c.Response().StatusCode()),
		zap.Int64("latency_ms", time.Since(start).Milliseconds()),
		zap.Any("request_id", c.Locals("request_id")),
		zap.String("ip", c.IP()),
		zap.String("user_agent", c.Get("User-Agent")),
	)

	return err
}

func RecoverMiddleware(c fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("panic recovered",
				zap.Any("error", r),
				zap.Any("request_id", c.Locals("request_id")),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Stack("stacktrace"),
			)
			_ = errorResponse(c, 500, "Internal Server Error")
		}
	}()
	return c.Next()
}
