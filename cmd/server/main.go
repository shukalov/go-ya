package main

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/server"
	serverlogger "github.com/shukalov/go-ya/internal/server/logger"
)

type envConfig struct {
	Address         *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDSN     *string `env:"DATABASE_DSN"`
	Key             *string `env:"KEY"`
}

func override[T comparable](flag *T, env *T) {
	if env != nil {
		*flag = *env
	}
}

func main() {
	logger := serverlogger.New()
	defer logger.Sync()

	addr := flag.String("a", "localhost:8080", "address endpoint")
	storeInterval := flag.Int("i", 300, "store interval in seconds (0 = sync)")
	fileStoragePath := flag.String("f", "/tmp/metrics-snapshot.json", "file storage path")
	restore := flag.Bool("r", false, "restore metrics from file on startup")
	databaseDSN := flag.String("d", "", "database DSN (PostgreSQL)")
	key := flag.String("k", "", "signing key")

	var envCfg envConfig
	if err := env.Parse(&envCfg); err != nil {
		logger.Fatal("failed to parse env", zap.Error(err))
	}

	if envCfg.Address == nil || envCfg.StoreInterval == nil || envCfg.FileStoragePath == nil || envCfg.DatabaseDSN == nil || envCfg.Key == nil {
		flag.Parse()
	}

	override(addr, envCfg.Address)
	override(storeInterval, envCfg.StoreInterval)
	override(fileStoragePath, envCfg.FileStoragePath)
	override(restore, envCfg.Restore)
	override(databaseDSN, envCfg.DatabaseDSN)
	override(key, envCfg.Key)

	srv := server.NewServer(server.Config{
		Address:       *addr,
		Logger:        logger,
		DatabaseDSN:   *databaseDSN,
		FilePath:      *fileStoragePath,
		StoreInterval: time.Duration(*storeInterval) * time.Second,
		Restore:       *restore,
		Key:           *key,
	})

	if err := srv.Run(); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
