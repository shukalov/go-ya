package agent

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/agent/collector"
	"github.com/shukalov/go-ya/internal/agent/sender"
	"github.com/shukalov/go-ya/pkg/models"
)

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
	Logger         *zap.Logger
	Key            string
	RateLimit      int
}

type Agent struct {
	config     Config
	collectors []collector.Collector
	sender     *sender.HTTPSender
	sends      chan models.RuntimeMetrics
	wg         sync.WaitGroup
}

func NewAgent(config Config, collectors []collector.Collector) *Agent {
	return &Agent{
		config:     config,
		collectors: collectors,
		sender:     sender.NewHTTPSender(config.ServerAddress, config.Key),
	}
}

func (a *Agent) startWorkers(ctx context.Context) {
	for i := 0; i < a.config.RateLimit; i++ {
		a.wg.Add(1)
		go func(id int) {
			defer a.wg.Done()
			for metrics := range a.sends {
				a.config.Logger.Info("worker sending metrics",
					zap.Int("workerID", id),
					zap.Int64("PollCount", metrics.Counters["PollCount"]),
				)
				if err := a.sender.SendAllMetrics(metrics); err != nil {
					a.config.Logger.Error("worker error sending metrics",
						zap.Int("workerID", id),
						zap.Error(err),
					)
				} else {
					a.config.Logger.Info("worker metrics sent",
						zap.Int("workerID", id),
						zap.Int64("PollCount", metrics.Counters["PollCount"]),
					)
				}
			}
		}(i)
	}
}

func (a *Agent) mergeMetrics() models.RuntimeMetrics {
	merged := models.RuntimeMetrics{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
	for _, c := range a.collectors {
		m := c.GetMetrics()
		for k, v := range m.Gauges {
			merged.Gauges[k] = v
		}
		for k, v := range m.Counters {
			merged.Counters[k] += v
		}
	}
	return merged
}

func (a *Agent) Run(ctx context.Context) {
	a.config.Logger.Info("starting agent",
		zap.Duration("pollInterval", a.config.PollInterval),
		zap.Duration("reportInterval", a.config.ReportInterval),
		zap.String("serverAddress", a.config.ServerAddress),
		zap.Int("rateLimit", a.config.RateLimit),
		zap.Int("collectors", len(a.collectors)),
	)

	a.sends = make(chan models.RuntimeMetrics, a.config.RateLimit)
	a.startWorkers(ctx)

	for _, c := range a.collectors {
		go c.Run(ctx)
	}

	reportTicker := time.NewTicker(a.config.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(a.sends)
			a.wg.Wait()
			a.config.Logger.Info("agent stopped", zap.Error(ctx.Err()))
			return
		case <-reportTicker.C:
			metrics := a.mergeMetrics()
			a.config.Logger.Info("metrics collected", zap.Int64("PollCount", metrics.Counters["PollCount"]))
			a.sends <- metrics
		}
	}
}
