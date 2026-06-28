package main

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/server"
	serverlogger "github.com/shukalov/go-ya/internal/server/logger"
	"github.com/shukalov/go-ya/internal/server/storage"
)

type envConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func override[T comparable](flag *T, env T) {
	var zero T
	if env != zero {
		*flag = env
	}
}

func main() {
	logger := serverlogger.New()
	defer logger.Sync()

	addr := flag.String("a", "localhost:8080", "address endpoint")
	storeInterval := flag.Int("i", 300, "store interval in seconds (0 = sync)")
	fileStoragePath := flag.String("f", "/tmp/metrics-snapshot.json", "file storage path")
	restore := flag.Bool("r", false, "restore metrics from file on startup")

	var envCfg envConfig
	if err := env.Parse(&envCfg); err != nil {
		logger.Fatal("failed to parse env", zap.Error(err))
	}

	if envCfg.Address == "" || envCfg.StoreInterval == 0 || envCfg.FileStoragePath == "" {
		flag.Parse()
	}

	override(addr, envCfg.Address)
	override(storeInterval, envCfg.StoreInterval)
	override(fileStoragePath, envCfg.FileStoragePath)
	override(restore, envCfg.Restore)

	fs := storage.NewFileBackedStorage(*fileStoragePath, time.Duration(*storeInterval)*time.Second)

	if *restore {
		if err := fs.Load(); err != nil {
			logger.Error("failed to restore metrics", zap.Error(err))
		}
	}

	if *storeInterval > 0 {
		go fs.Run()
	}

	srv := server.NewServer(*addr, fs, logger)

	if err := srv.Run(); err != nil {
		logger.Fatal("server error", zap.Error(err))
	}
}
