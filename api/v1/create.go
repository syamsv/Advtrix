package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/syamsv/Advtrix/common/mongodb"
	"github.com/syamsv/Advtrix/common/redis"
	"github.com/syamsv/Advtrix/common/totp"
	"github.com/syamsv/Advtrix/common/views"
	"github.com/syamsv/Advtrix/models"
)

const (
	secureCodeCollection = "secure_codes"
	secureCodeCacheTTL   = 24 * time.Hour
)

// Generate handles POST requests to create a new SecureCode with a random TOTP secret.
func Create(c fiber.Ctx) error {
	var req models.SecureCode
	if err := c.Bind().JSON(&req); err != nil {
		return views.InvalidParams(c)
	}

	secret, err := totp.GenerateSecret(20)
	if err != nil {
		return views.InternalServerError(c, err)
	}

	sc := models.SecureCode{
		Id:       req.Id,
		Metadata: req.Metadata,
		Secret:   secret,
	}

	if err := createSecureCode(c.Context(), sc); err != nil {
		return views.InternalServerError(c, err)
	}

	return views.CreatedResponse(c, fmt.Sprintf("secret created for %s", sc.Id))
}

// --- data controller ---

func createSecureCode(ctx context.Context, sc models.SecureCode) error {
	if _, err := mongodb.InsertOne(ctx, secureCodeCollection, sc); err != nil {
		return err
	}

	data, err := json.Marshal(sc)
	if err != nil {
		zap.L().Error("failed to marshal secure code for cache", zap.Error(err))
		return nil
	}

	if err := redis.SetJSON(ctx, cacheKey(sc.Id), data, secureCodeCacheTTL); err != nil {
		zap.L().Error("failed to cache secure code", zap.String("id", sc.Id), zap.Error(err))
	}

	return nil
}

func cacheKey(id string) string {
	return "secure_code:" + id
}
