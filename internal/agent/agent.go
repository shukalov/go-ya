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

	// Первый сбор метрик
	a.collector.Collect()
	fmt.Println("Initial metrics collected")

	reportCounter := 0
	reportsPerCycle := int(a.config.ReportInterval / a.config.PollInterval)

	for {
		// Собираем метрики
		a.collector.Collect()
		pollCount := a.collector.GetPollCount()
		fmt.Printf("Metrics collected (PollCount: %d)\n", pollCount)

		// Проверяем, нужно ли отправлять
		reportCounter++
		if reportCounter >= reportsPerCycle {
			metrics := a.collector.GetMetrics()
			if err := a.sender.SendAllMetrics(metrics, pollCount); err != nil {
				fmt.Printf("Error sending metrics: %v\n", err)
			} else {
				fmt.Printf("Metrics sent successfully (PollCount: %d)\n", pollCount)
			}
			reportCounter = 0
		}

		time.Sleep(a.config.PollInterval)
	}
}
