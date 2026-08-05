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
	Address        *string `env:"ADDRESS"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
	Key            *string `env:"KEY"`
}

func override[T comparable](flag *T, env *T) {
	if env != nil {
		*flag = *env
	}
}

func main() {
	logger := agentlogger.New()
	defer logger.Sync()

	addr := flag.String("a", "localhost:8080", "address endpoint")
	reportInterval := flag.Int("r", 10, "report interval in seconds")
	pollInterval := flag.Int("p", 2, "poll interval in seconds")
	key := flag.String("k", "", "signing key")

	var envCfg envConfig
	if err := env.Parse(&envCfg); err != nil {
		logger.Fatal("failed to parse env", zap.Error(err))
	}

	if envCfg.Address == nil || envCfg.ReportInterval == nil || envCfg.PollInterval == nil || envCfg.Key == nil {
		flag.Parse()
	}

	override(addr, envCfg.Address)
	override(reportInterval, envCfg.ReportInterval)
	override(pollInterval, envCfg.PollInterval)
	override(key, envCfg.Key)

	if *reportInterval <= 0 || *pollInterval <= 0 {
		logger.Fatal("intervals must be positive")
	}

	agentConfig := agent.Config{
		PollInterval:   time.Duration(*pollInterval) * time.Second,
		ReportInterval: time.Duration(*reportInterval) * time.Second,
		ServerAddress:  "http://" + *addr,
		Logger:         logger,
		Key:            *key,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a := agent.NewAgent(agentConfig)
	a.Run(ctx)
}
