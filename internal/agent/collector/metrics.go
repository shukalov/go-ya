package collector

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/shukalov/go-ya/pkg/models"
)

// MetricsCollector - сборщик метрик из runtime
type MetricsCollector struct {
	mu        sync.RWMutex
	metrics   *models.RuntimeMetrics
	pollCount int64
}

// NewMetricsCollector - создает новый сборщик
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:   &models.RuntimeMetrics{},
		pollCount: 0,
	}
}

// Collect - собирает метрики из runtime
func (c *MetricsCollector) Collect() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Заполняем gauge метрики
	c.metrics.Alloc = float64(memStats.Alloc)
	c.metrics.BuckHashSys = float64(memStats.BuckHashSys)
	c.metrics.Frees = float64(memStats.Frees)
	c.metrics.GCCPUFraction = memStats.GCCPUFraction
	c.metrics.GCSys = float64(memStats.GCSys)
	c.metrics.HeapAlloc = float64(memStats.HeapAlloc)
	c.metrics.HeapIdle = float64(memStats.HeapIdle)
	c.metrics.HeapInuse = float64(memStats.HeapInuse)
	c.metrics.HeapObjects = float64(memStats.HeapObjects)
	c.metrics.HeapReleased = float64(memStats.HeapReleased)
	c.metrics.HeapSys = float64(memStats.HeapSys)
	c.metrics.LastGC = float64(memStats.LastGC)
	c.metrics.Lookups = float64(memStats.Lookups)
	c.metrics.MCacheInuse = float64(memStats.MCacheInuse)
	c.metrics.MCacheSys = float64(memStats.MCacheSys)
	c.metrics.MSpanInuse = float64(memStats.MSpanInuse)
	c.metrics.MSpanSys = float64(memStats.MSpanSys)
	c.metrics.Mallocs = float64(memStats.Mallocs)
	c.metrics.NextGC = float64(memStats.NextGC)
	c.metrics.NumForcedGC = float64(memStats.NumForcedGC)
	c.metrics.NumGC = float64(memStats.NumGC)
	c.metrics.OtherSys = float64(memStats.OtherSys)
	c.metrics.PauseTotalNs = float64(memStats.PauseTotalNs)
	c.metrics.StackInuse = float64(memStats.StackInuse)
	c.metrics.StackSys = float64(memStats.StackSys)
	c.metrics.Sys = float64(memStats.Sys)
	c.metrics.TotalAlloc = float64(memStats.TotalAlloc)

	// Добавляем RandomValue
	c.metrics.RandomValue = rand.Float64() * 100

	// Увеличиваем счетчик
	c.pollCount++
	c.metrics.PollCount = c.pollCount
}

// GetMetrics - возвращает копию собранных метрик
func (c *MetricsCollector) GetMetrics() models.RuntimeMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c.metrics
}

// GetPollCount - возвращает текущее значение счетчика
func (c *MetricsCollector) GetPollCount() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pollCount
}
