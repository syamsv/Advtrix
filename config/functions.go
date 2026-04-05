package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// ENV lists environment variable names that are required to be non-empty.
var ENV = []string{}

func LoadConfig(filename, filetype, path string) {
	importConfig(filename, filetype, path)

	APP_NAME = viper.GetString("APP_NAME")
	ENVIRONMENT = viper.GetString("ENVIRONMENT")

	SERVER_ADDRESS = viper.GetString("SERVER_ADDRESS")
	INTERNAL_AUTH_PARAMATER = viper.GetString("INTERNAL_AUTH_PARAMATER")

	MONGODB_NAME = viper.GetString("MONGODB_NAME")
	MONGODB_HOST = viper.GetString("MONGODB_HOST")
	MONGODB_PORT = viper.GetString("MONGODB_PORT")
	MONGO_ROOT_USER = viper.GetString("MONGO_ROOT_USER")
	MONGO_ROOT_PASSWORD = viper.GetString("MONGO_ROOT_PASSWORD")
	MONGODB_URI = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
		MONGO_ROOT_USER,
		MONGO_ROOT_PASSWORD,
		MONGODB_HOST,
		MONGODB_PORT,
		MONGODB_NAME,
	)

	REDIS_ADDRESS = viper.GetString("REDIS_ADDRESS")
	REDIS_PASSWORD = viper.GetString("REDIS_PASSWORD")
	REDIS_DB = viper.GetInt("REDIS_DB")

	NTS_SERVER = viper.GetString("NTS_SERVER")
	STEPSECOND = viper.GetInt("STEPSECOND")
	ENCRYPTION_KEY = viper.GetString("ENCRYPTION_KEY")

	validate()
}

func importConfig(filename, filetype, path string) {
	viper.SetConfigName(filename)
	viper.SetConfigType(filetype)
	viper.AddConfigPath(path)

	// Application defaults.
	viper.SetDefault("APP_NAME", "Advtrix")
	viper.SetDefault("APP_VERSION", "0.1.0")
	viper.SetDefault("ENVIRONMENT", "development")

	// Server defaults.
	viper.SetDefault("SERVER_ADDRESS", ":9000")

	// Database defaults.
	viper.SetDefault("MONGODB_HOST", "localhost")
	viper.SetDefault("MONGODB_PORT", "27017")
	viper.SetDefault("MONGODB_NAME", "advtrix")
	viper.SetDefault("MONGO_ROOT_USER", "root")
	viper.SetDefault("MONGO_ROOT_PASSWORD", "secret")
	viper.SetDefault("REDIS_ADDRESS", "localhost:6379")
	viper.SetDefault("REDIS_DB", 0)

	// NTS / TOTP defaults.
	viper.SetDefault("NTS_SERVER", "time.cloudflare.com")
	viper.SetDefault("STEPSECOND", 15)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Panicln(fmt.Errorf("fatal error config file: %s", err))
		}
	}

	for _, element := range ENV {
		if viper.GetString(element) == "" {
			log.Panicln(fmt.Errorf("env variables not present %s", element))
		}
	}
}

func validate() {
	if STEPSECOND <= 0 {
		log.Panicln("STEPSECOND must be a positive integer")
	}

	if NTS_SERVER == "" {
		log.Panicln("NTS_SERVER must not be empty")
	}

	if SERVER_ADDRESS == "" {
		log.Panicln("SERVER_ADDRESS must not be empty")
	}

	if MONGODB_HOST == "" {
		log.Panicln("MONGODB_HOST must not be empty")
	}

	if MONGODB_PORT == "" {
		log.Panicln("MONGODB_PORT must not be empty")
	}

	if INTERNAL_AUTH_PARAMATER == "" {
		log.Panicln("INTERNAL_AUTH_PARAMATER must not be empty — set it to a strong random token")
	}

	if len(ENCRYPTION_KEY) != 64 {
		log.Panicln("ENCRYPTION_KEY must be a 64-character hex string (32 bytes for AES-256)")
	}
}
