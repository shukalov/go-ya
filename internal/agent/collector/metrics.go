package collector

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/shukalov/go-ya/pkg/models"
)

// MetricsCollector - сборщик метрик из runtime
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics models.RuntimeMetrics
}

// NewMetricsCollector - создает новый сборщик
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: models.RuntimeMetrics{
			Gauges:   make(map[string]float64, len(models.GaugeDefs())),
			Counters: make(map[string]int64, len(models.CounterDefs())),
		},
	}
}

// Collect - собирает метрики из runtime
func (c *MetricsCollector) Collect() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.mu.Lock()
	defer c.mu.Unlock()

	g := c.metrics.Gauges

	g["Alloc"] = float64(memStats.Alloc)
	g["BuckHashSys"] = float64(memStats.BuckHashSys)
	g["Frees"] = float64(memStats.Frees)
	g["GCCPUFraction"] = memStats.GCCPUFraction
	g["GCSys"] = float64(memStats.GCSys)
	g["HeapAlloc"] = float64(memStats.HeapAlloc)
	g["HeapIdle"] = float64(memStats.HeapIdle)
	g["HeapInuse"] = float64(memStats.HeapInuse)
	g["HeapObjects"] = float64(memStats.HeapObjects)
	g["HeapReleased"] = float64(memStats.HeapReleased)
	g["HeapSys"] = float64(memStats.HeapSys)
	g["LastGC"] = float64(memStats.LastGC)
	g["Lookups"] = float64(memStats.Lookups)
	g["MCacheInuse"] = float64(memStats.MCacheInuse)
	g["MCacheSys"] = float64(memStats.MCacheSys)
	g["MSpanInuse"] = float64(memStats.MSpanInuse)
	g["MSpanSys"] = float64(memStats.MSpanSys)
	g["Mallocs"] = float64(memStats.Mallocs)
	g["NextGC"] = float64(memStats.NextGC)
	g["NumForcedGC"] = float64(memStats.NumForcedGC)
	g["NumGC"] = float64(memStats.NumGC)
	g["OtherSys"] = float64(memStats.OtherSys)
	g["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	g["StackInuse"] = float64(memStats.StackInuse)
	g["StackSys"] = float64(memStats.StackSys)
	g["Sys"] = float64(memStats.Sys)
	g["TotalAlloc"] = float64(memStats.TotalAlloc)

	g["RandomValue"] = rand.Float64() * 100

	c.metrics.Counters["PollCount"]++
}

// GetMetrics - возвращает копию собранных метрик
func (c *MetricsCollector) GetMetrics() models.RuntimeMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	gauges := make(map[string]float64, len(c.metrics.Gauges))
	for k, v := range c.metrics.Gauges {
		gauges[k] = v
	}
	counters := make(map[string]int64, len(c.metrics.Counters))
	for k, v := range c.metrics.Counters {
		counters[k] = v
	}
	return models.RuntimeMetrics{Gauges: gauges, Counters: counters}
}
