package v1

import (
	"crypto/sha256"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/syamsv/Advtrix/common/totp"
	"github.com/syamsv/Advtrix/common/views"
)

type getCodeRequest struct {
	Id string `json:"id"`
}

// GetCode handles POST /getcode.
// Looks up the SecureCode, generates a TOTP from its secret, and returns
// sha256(sha256(id):sha256(totp)).
func GetCode(c fiber.Ctx) error {
	var req getCodeRequest
	if err := c.Bind().JSON(&req); err != nil || req.Id == "" {
		return views.InvalidParams(c)
	}

	sc, err := getSecureCode(c.Context(), req.Id)
	if err != nil {
		return views.RecordNotFound(c)
	}

	secret, err := decryptSecret(sc)
	if err != nil {
		return views.InternalServerError(c, err)
	}

	code, remaining, err := totp.GenerateTOTP(secret)
	if err != nil {
		return views.InternalServerError(c, err)
	}

	hash := hashCode(req.Id, code)

	return views.SuccessResponse(c, fiber.Map{
		"id":        req.Id,
		"code":      hash,
		"remaining": remaining,
	})
}

// hashCode produces sha256(sha256(id):sha256(totp)).
func hashCode(id, code string) string {
	idHash := fmt.Sprintf("%x", sha256.Sum256([]byte(id)))
	codeHash := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))
	combined := fmt.Sprintf("%s:%s", idHash, codeHash)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(combined)))
}
