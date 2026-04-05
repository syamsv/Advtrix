package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/syamsv/Advtrix/common/crypto"
	"github.com/syamsv/Advtrix/common/logger"
	"github.com/syamsv/Advtrix/common/mongodb"
	"github.com/syamsv/Advtrix/common/nts"
	"github.com/syamsv/Advtrix/common/redis"
	"github.com/syamsv/Advtrix/config"
	"github.com/syamsv/Advtrix/server"

	v1 "github.com/syamsv/Advtrix/api/v1"

	"go.uber.org/zap"
)

func main() {
	config.LoadConfig(".env", "env", ".")
	logger.Init()
	defer logger.Sync()

	crypto.Init(config.ENCRYPTION_KEY)
	mongodb.Init()
	redis.Init()
	nts.Init()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := v1.EnsureIndexes(ctx); err != nil {
		zap.L().Fatal("failed to create indexes", zap.Error(err))
	}

	go server.Run(config.SERVER_ADDRESS)

	zap.L().Info("server started",
		zap.String("address", config.SERVER_ADDRESS),
		zap.String("env", config.ENVIRONMENT),
		zap.String("app", config.APP_NAME),
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
