package agent

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/agent/collector"
	"github.com/shukalov/go-ya/internal/agent/sender"
)

// Config - конфигурация агента
type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
	Logger         *zap.Logger
	Key            string
}

// Agent - основной компонент агента
type Agent struct {
	config    Config
	collector *collector.MetricsCollector
	sender    *sender.HTTPSender
}

// NewAgent - создает нового агента
func NewAgent(config Config) *Agent {
	return &Agent{
		config:    config,
		collector: collector.NewMetricsCollector(),
		sender:    sender.NewHTTPSender(config.ServerAddress, config.Key),
	}
}

// Run - запускает агента, блокируется до отмены контекста
func (a *Agent) Run(ctx context.Context) {
	a.config.Logger.Info("starting agent",
		zap.Duration("pollInterval", a.config.PollInterval),
		zap.Duration("reportInterval", a.config.ReportInterval),
		zap.String("serverAddress", a.config.ServerAddress),
	)

	a.collector.Collect()

	pollTicker := time.NewTicker(a.config.PollInterval)
	reportTicker := time.NewTicker(a.config.ReportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.config.Logger.Info("agent stopped", zap.Error(ctx.Err()))
			return
		case <-pollTicker.C:
			a.collector.Collect()
		case <-reportTicker.C:
			metrics := a.collector.GetMetrics()
			a.config.Logger.Info("metrics collected", zap.Int64("PollCount", metrics.Counters["PollCount"]))

			if err := a.sender.SendAllMetrics(metrics); err != nil {
				a.config.Logger.Error("error sending metrics", zap.Error(err))
			} else {
				a.config.Logger.Info("metrics sent", zap.Int64("PollCount", metrics.Counters["PollCount"]))
			}
		}
	}
}
