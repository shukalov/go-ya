package sys

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"

	"github.com/shukalov/go-ya/pkg/models"
)

type SysCollector struct {
	mu           sync.RWMutex
	gauges       map[string]float64
	pollInterval time.Duration
	logger       *zap.Logger
	numCPU       int
}

func New(pollInterval time.Duration, logger *zap.Logger) *SysCollector {
	return &SysCollector{
		gauges:       make(map[string]float64),
		pollInterval: pollInterval,
		logger:       logger,
		numCPU:       runtime.NumCPU(),
	}
}

func (c *SysCollector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	vmem, err := mem.VirtualMemory()
	if err != nil {
		c.logger.Error("failed to get virtual memory", zap.Error(err))
		return
	}
	c.gauges["TotalMemory"] = float64(vmem.Total)
	c.gauges["FreeMemory"] = float64(vmem.Free)

	percents, err := cpu.Percent(time.Second, true)
	if err != nil {
		c.logger.Error("failed to get cpu percent", zap.Error(err))
		return
	}
	for i, p := range percents {
		c.gauges[fmt.Sprintf("CPUutilization1_%d", i)] = p
	}
}

func (c *SysCollector) GetMetrics() models.RuntimeMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	gauges := make(map[string]float64, len(c.gauges))
	for k, v := range c.gauges {
		gauges[k] = v
	}
	return models.RuntimeMetrics{Gauges: gauges, Counters: make(map[string]int64)}
}

func (c *SysCollector) Run(ctx context.Context) {
	c.Collect()

	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	c.logger.Info("sys collector started", zap.Duration("pollInterval", c.pollInterval))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("sys collector stopped")
			return
		case <-ticker.C:
			c.Collect()
		}
	}
}
