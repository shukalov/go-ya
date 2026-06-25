package agent

import (
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

// Run - запускает агента
func (a *Agent) Run() error {
	fmt.Printf("Starting agent with pollInterval=%v, reportInterval=%v\n",
		a.config.PollInterval, a.config.ReportInterval)
	fmt.Printf("Server address: %s\n", a.config.ServerAddress)

	a.collector.Collect()

	reportCounter := 0
	reportsPerCycle := int(a.config.ReportInterval / a.config.PollInterval)

	for {
		a.collector.Collect()

		reportCounter++
		if reportCounter >= reportsPerCycle {
			metrics := a.collector.GetMetrics()
			fmt.Printf("Metrics collected (PollCount: %d)\n", metrics.Counters["PollCount"])

			if err := a.sender.SendAllMetrics(metrics); err != nil {
				fmt.Printf("Error sending metrics: %v\n", err)
			} else {
				fmt.Printf("Metrics sent successfully (PollCount: %d)\n", metrics.Counters["PollCount"])
			}
			reportCounter = 0
		}

		time.Sleep(a.config.PollInterval)
	}
}
