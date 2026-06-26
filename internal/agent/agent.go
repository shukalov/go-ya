package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/shukalov/go-ya/internal/agent/collector"
	"github.com/shukalov/go-ya/internal/agent/sender"
)

// Config - конфигурация агента
type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
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
		sender:    sender.NewHTTPSender(config.ServerAddress),
	}
}

// Run - запускает агента, блокируется до отмены контекста
func (a *Agent) Run(ctx context.Context) error {
	fmt.Printf("Starting agent with pollInterval=%v, reportInterval=%v\n",
		a.config.PollInterval, a.config.ReportInterval)
	fmt.Printf("Server address: %s\n", a.config.ServerAddress)

	a.collector.Collect()

	pollTicker := time.NewTicker(a.config.PollInterval)
	reportTicker := time.NewTicker(a.config.ReportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Agent stopped: %v\n", ctx.Err())
			return nil
		case <-pollTicker.C:
			a.collector.Collect()
		case <-reportTicker.C:
			metrics := a.collector.GetMetrics()
			fmt.Printf("Metrics collected (PollCount: %d)\n", metrics.Counters["PollCount"])

			if err := a.sender.SendAllMetrics(metrics); err != nil {
				fmt.Printf("Error sending metrics: %v\n", err)
			} else {
				fmt.Printf("Metrics sent successfully (PollCount: %d)\n", metrics.Counters["PollCount"])
			}
		}
	}
}

