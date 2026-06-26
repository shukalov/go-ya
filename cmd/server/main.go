package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/shukalov/go-ya/internal/server"
	"github.com/shukalov/go-ya/internal/server/storage"
)

type envConfig struct {
	Address string `env:"ADDRESS"`
}

func main() {
	addr := flag.String("a", "localhost:8080", "address endpoint")

	var envCfg envConfig
	if err := env.Parse(&envCfg); err != nil {
		log.Fatalf("Failed to parse env: %v\n", err)
	}

	if envCfg.Address == "" {
		flag.Parse()
	}

	serverAddress := envCfg.Address
	if serverAddress == "" {
		serverAddress = *addr
	}

	storage := storage.NewMemStorage()
	srv := server.NewServer(serverAddress, storage)

	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
