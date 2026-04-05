package v1

import (
	"crypto/hmac"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/syamsv/Advtrix/common/ratelimit"
	"github.com/syamsv/Advtrix/common/totp"
	"github.com/syamsv/Advtrix/common/views"
	"github.com/syamsv/Advtrix/config"
)

// validateLimiter allows 10 validation attempts per ID per 30-second window.
var validateLimiter = ratelimit.New(10, 30*time.Second)

type validateRequest struct {
	Id   string `json:"id"`
	Code string `json:"code"`
}

// Validate handles POST /validate.
// Regenerates TOTP codes for the current and adjacent time steps,
// hashes each the same way GetCode does, and checks for a match.
func Validate(c fiber.Ctx) error {
	var req validateRequest
	if err := c.Bind().JSON(&req); err != nil || req.Id == "" || req.Code == "" {
		return views.InvalidParams(c)
	}

	if !validateLimiter.Allow(req.Id) {
		return views.ErrorResponse(c, 429, "Too many validation attempts. Try again later.")
	}

	sc, err := getSecureCode(c.Context(), req.Id)
	if err != nil {
		return views.RecordNotFound(c)
	}

	secret, err := decryptSecret(sc)
	if err != nil {
		return views.InternalServerError(c, err)
	}

	valid, err := validateHash(req.Id, req.Code, secret)
	if err != nil {
		return views.InternalServerError(c, err)
	}

	return views.SuccessResponse(c, fiber.Map{
		"id":    req.Id,
		"valid": valid,
	})
}

// validateHash checks the provided hash against TOTP codes for the current
// and ±1 adjacent time steps.
func validateHash(id, providedHash, secret string) (bool, error) {
	now := totp.NTSNow()
	step := uint64(config.STEPSECOND)
	currentCounter := uint64(now.Unix()) / step

	for _, counter := range []uint64{currentCounter - 1, currentCounter, currentCounter + 1} {
		code, err := totp.GenerateTOTPAtCounter(secret, counter)
		if err != nil {
			return false, err
		}

		expected := hashCode(id, code)
		if hmac.Equal([]byte(expected), []byte(providedHash)) {
			return true, nil
		}
	}

	return false, nil
}
