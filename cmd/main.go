package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/syamsv/Advtrix/common/logger"
	"github.com/syamsv/Advtrix/common/mongodb"
	"github.com/syamsv/Advtrix/common/nts"
	"github.com/syamsv/Advtrix/common/redis"
	"github.com/syamsv/Advtrix/config"
	"github.com/syamsv/Advtrix/server"
	"go.uber.org/zap"
)

func main() {
	config.LoadConfig(".env", "env", ".")
	logger.Init()
	defer logger.Sync()

	mongodb.Init()
	redis.Init()
	nts.Init()

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
	nts.Shutdown()
	mongodb.Close()
	redis.Close()

	zap.L().Info("shutdown complete")
}
