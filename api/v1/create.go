package v1

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/syamsv/Advtrix/common/crypto"
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

// EnsureIndexes creates required MongoDB indexes. Call once at startup.
func EnsureIndexes(ctx context.Context) error {
	return mongodb.CreateUniqueIndex(ctx, secureCodeCollection, "id")
}

// Create handles POST requests to create a new SecureCode with a random TOTP secret.
func Create(c fiber.Ctx) error {
	var req models.SecureCode
	if err := c.Bind().JSON(&req); err != nil {
		return views.InvalidParams(c)
	}

	if err := validateInput(req); err != nil {
		return views.ErrorResponse(c, 400, err.Error())
	}

	secret, err := totp.GenerateSecret(20)
	if err != nil {
		return views.InternalServerError(c, err)
	}

	encrypted, err := crypto.Encrypt(secret)
	if err != nil {
		return views.InternalServerError(c, err)
	}

	sc := models.SecureCode{
		Id:              req.Id,
		Metadata:        req.Metadata,
		EncryptedSecret: encrypted,
	}

	if err := createSecureCode(c.Context(), sc); err != nil {
		return views.ErrorResponse(c, 409, "entry with this ID already exists")
	}

	return views.CreatedResponse(c, fmt.Sprintf("secret created for %s", sc.Id))
}

func validateInput(req models.SecureCode) error {
	if req.Id == "" {
		return fmt.Errorf("id must not be empty")
	}
	if utf8.RuneCountInString(req.Id) > models.MaxIDLength {
		return fmt.Errorf("id must not exceed %d characters", models.MaxIDLength)
	}
	if req.Metadata != nil {
		data, err := json.Marshal(req.Metadata)
		if err != nil {
			return fmt.Errorf("metadata is not valid JSON")
		}
		if len(data) > models.MaxMetadataBytes {
			return fmt.Errorf("metadata must not exceed %d bytes", models.MaxMetadataBytes)
		}
	}
	return nil
}

// --- data controller ---

func createSecureCode(ctx context.Context, sc models.SecureCode) error {
	if _, err := mongodb.InsertOne(ctx, secureCodeCollection, sc); err != nil {
		return err
	}

	cacheSecureCode(ctx, sc)
	return nil
}

func getSecureCode(ctx context.Context, id string) (models.SecureCode, error) {
	var sc models.SecureCode

	// Try cache first.
	data, err := redis.GetJSON(ctx, cacheKey(id))
	if err == nil {
		var cached models.SecureCodeCache
		if err := json.Unmarshal(data, &cached); err == nil {
			return cached.ToSecureCode(), nil
		}
	}

	// Fallback to database.
	filter := bson.M{"id": id}
	if err := mongodb.FindOne(ctx, secureCodeCollection, filter, &sc); err != nil {
		return sc, err
	}

	// Backfill cache.
	cacheSecureCode(ctx, sc)

	return sc, nil
}

// decryptSecret decrypts the EncryptedSecret field and returns the plaintext TOTP secret.
func decryptSecret(sc models.SecureCode) (string, error) {
	return crypto.Decrypt(sc.EncryptedSecret)
}

func cacheSecureCode(ctx context.Context, sc models.SecureCode) {
	raw, err := json.Marshal(sc.ToCache())
	if err != nil {
		zap.L().Error("failed to marshal secure code for cache", zap.Error(err))
		return
	}
	if err := redis.SetJSON(ctx, cacheKey(sc.Id), raw, secureCodeCacheTTL); err != nil {
		zap.L().Error("failed to cache secure code", zap.String("id", sc.Id), zap.Error(err))
	}
}

// cacheKey produces a sanitised Redis key to prevent injection.
func cacheKey(id string) string {
	return "secure_code:" + hex.EncodeToString([]byte(id))
}
