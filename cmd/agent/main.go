package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/agent"
	agentlogger "github.com/shukalov/go-ya/internal/agent/logger"
)

type envConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func main() {
	logger := agentlogger.New()
	defer logger.Sync()

	addr := flag.String("a", "localhost:8080", "address endpoint")
	reportInterval := flag.Int("r", 10, "report interval in seconds")
	pollInterval := flag.Int("p", 2, "poll interval in seconds")

	var envCfg envConfig
	if err := env.Parse(&envCfg); err != nil {
		logger.Fatal("failed to parse env", zap.Error(err))
	}

	if envCfg.Address == "" || envCfg.ReportInterval == 0 || envCfg.PollInterval == 0 {
		flag.Parse()
	}

	serverAddress := envCfg.Address
	if serverAddress == "" {
		serverAddress = *addr
	}

	reportIntervalSec := envCfg.ReportInterval
	if reportIntervalSec == 0 {
		reportIntervalSec = *reportInterval
	}

	pollIntervalSec := envCfg.PollInterval
	if pollIntervalSec == 0 {
		pollIntervalSec = *pollInterval
	}

	if reportIntervalSec <= 0 || pollIntervalSec <= 0 {
		logger.Fatal("intervals must be positive")
	}

	config := agent.Config{
		PollInterval:   time.Duration(pollIntervalSec) * time.Second,
		ReportInterval: time.Duration(reportIntervalSec) * time.Second,
		ServerAddress:  "http://" + serverAddress,
		Logger:         logger,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a := agent.NewAgent(config)
	a.Run(ctx)
}
