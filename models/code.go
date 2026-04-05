package models

const (
	MaxIDLength       = 256
	MaxMetadataBytes  = 4096
)

type SecureCode struct {
	Id              string `json:"id" bson:"id"`
	Metadata        any    `json:"metadata" bson:"metadata"`
	EncryptedSecret string `json:"-" bson:"encrypted_secret"`
}

// SecureCodeCache is used for Redis serialization so the secret is included.
type SecureCodeCache struct {
	Id              string `json:"id"`
	Metadata        any    `json:"metadata"`
	EncryptedSecret string `json:"encrypted_secret"`
}

func (sc SecureCode) ToCache() SecureCodeCache {
	return SecureCodeCache{
		Id:              sc.Id,
		Metadata:        sc.Metadata,
		EncryptedSecret: sc.EncryptedSecret,
	}
}

func (c SecureCodeCache) ToSecureCode() SecureCode {
	return SecureCode{
		Id:              c.Id,
		Metadata:        c.Metadata,
		EncryptedSecret: c.EncryptedSecret,
	}
}
