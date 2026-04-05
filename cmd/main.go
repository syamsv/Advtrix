package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/syamsv/go-template/common/logger"
	"github.com/syamsv/go-template/common/mongodb"
	"github.com/syamsv/go-template/common/redis"
	"github.com/syamsv/go-template/config"
	"github.com/syamsv/go-template/server"
)

func init() {
	config.LoadConfig(".env", "env", ".")
	logger.Init()
}

func main() {
	defer logger.Sync()

	mongodb.Init()
	redis.Init()

	go server.Run(config.SERVER_ADDRESS)

	zap.L().Info("server started",
		zap.String("address", config.SERVER_ADDRESS),
		zap.String("env", config.ENVIRONMENT),
		zap.String("app", config.APP_NAME),
		zap.String("version", config.APP_VERSION),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("shutting down")

	server.Shutdown()
	mongodb.Close()
	redis.Close()

	zap.L().Info("shutdown complete")
}
