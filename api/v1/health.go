package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/syamsv/Advtrix/common/nts"
	"github.com/syamsv/Advtrix/common/views"
)

func Health(c fiber.Ctx) error {
	if !nts.Healthy() {
		return views.ErrorResponse(c, 503, "DEGRADED")
	}

	return views.SuccessResponse(c, fiber.Map{
		"status":     "READY",
		"nts_offset": nts.Offset().String(),
	})
}
