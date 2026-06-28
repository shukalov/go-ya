package main

import (
	"flag"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/server"
	serverlogger "github.com/shukalov/go-ya/internal/server/logger"
	"github.com/shukalov/go-ya/internal/server/storage"
)

type envConfig struct {
	Address string `env:"ADDRESS"`
}

func main() {
	logger := serverlogger.New()
	defer logger.Sync()

	addr := flag.String("a", "localhost:8080", "address endpoint")

	var envCfg envConfig
	if err := env.Parse(&envCfg); err != nil {
		logger.Fatal("failed to parse env", zap.Error(err))
	}

	if envCfg.Address == "" {
		flag.Parse()
	}

	serverAddress := envCfg.Address
	if serverAddress == "" {
		serverAddress = *addr
	}

	storage := storage.NewMemStorage()
	srv := server.NewServer(serverAddress, storage, logger)

	if err := srv.Run(); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
