package main

import (
	"context"
	"flag"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/agent"
	"github.com/shukalov/go-ya/internal/agent/collector"
	runtimecollector "github.com/shukalov/go-ya/internal/agent/collector/runtime"
	"github.com/shukalov/go-ya/internal/agent/collector/sys"
	agentlogger "github.com/shukalov/go-ya/internal/agent/logger"
)

type envConfig struct {
	Address        *string `env:"ADDRESS"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
	Key            *string `env:"KEY"`
	RateLimit      *int    `env:"RATE_LIMIT"`
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
	rateLimit := flag.Int("l", runtime.NumCPU(), "rate limit for concurrent requests")

	var envCfg envConfig
	if err := env.Parse(&envCfg); err != nil {
		logger.Fatal("failed to parse env", zap.Error(err))
	}

	if envCfg.Address == nil || envCfg.ReportInterval == nil || envCfg.PollInterval == nil || envCfg.Key == nil || envCfg.RateLimit == nil {
		flag.Parse()
	}

	override(addr, envCfg.Address)
	override(reportInterval, envCfg.ReportInterval)
	override(pollInterval, envCfg.PollInterval)
	override(key, envCfg.Key)
	override(rateLimit, envCfg.RateLimit)

	if *reportInterval <= 0 || *pollInterval <= 0 {
		logger.Fatal("intervals must be positive")
	}

	pollDur := time.Duration(*pollInterval) * time.Second

	collectors := []collector.Collector{
		runtimecollector.New(pollDur, logger),
		sys.New(pollDur, logger),
	}

	agentConfig := agent.Config{
		PollInterval:   pollDur,
		ReportInterval: time.Duration(*reportInterval) * time.Second,
		ServerAddress:  "http://" + *addr,
		Logger:         logger,
		Key:            *key,
		RateLimit:      *rateLimit,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a := agent.NewAgent(agentConfig, collectors)
	a.Run(ctx)
}
