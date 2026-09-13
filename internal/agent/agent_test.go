package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/agent/collector"
	"github.com/shukalov/go-ya/pkg/models"
)

type mockCollector struct {
	mu      sync.RWMutex
	metrics models.RuntimeMetrics
	delay   time.Duration
}

func newMockCollector() *mockCollector {
	return &mockCollector{
		metrics: models.RuntimeMetrics{
			Gauges:   map[string]float64{"TestGauge": 1.0},
			Counters: map[string]int64{"PollCount": 0},
		},
	}
}

func (m *mockCollector) Run(ctx context.Context) {
	m.Collect()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.Collect()
		}
	}
}

func (m *mockCollector) Collect() {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.Counters["PollCount"]++
}

func (m *mockCollector) GetMetrics() models.RuntimeMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	gauges := make(map[string]float64, len(m.metrics.Gauges))
	for k, v := range m.metrics.Gauges {
		gauges[k] = v
	}
	counters := make(map[string]int64, len(m.metrics.Counters))
	for k, v := range m.metrics.Counters {
		counters[k] = v
	}
	return models.RuntimeMetrics{Gauges: gauges, Counters: counters}
}

var _ collector.Collector = (*mockCollector)(nil)

func TestNewAgent(t *testing.T) {
	config := Config{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerAddress:  "http://localhost:8080",
		Logger:         zap.NewNop(),
		RateLimit:      2,
	}

	mc := newMockCollector()
	a := NewAgent(config, []collector.Collector{mc})
	if a == nil {
		t.Error("expected agent to be created")
	}
	if a.config.PollInterval != 2*time.Second {
		t.Errorf("expected PollInterval 2s, got %v", a.config.PollInterval)
	}
	if len(a.collectors) != 1 {
		t.Errorf("expected 1 collector, got %d", len(a.collectors))
	}
	if a.sender == nil {
		t.Error("expected sender to be initialized")
	}
}

func TestAgent_CollectMetrics(t *testing.T) {
	config := Config{
		PollInterval:   1 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerAddress:  "http://localhost:8080",
		Logger:         zap.NewNop(),
		RateLimit:      2,
	}

	mc := newMockCollector()
	_ = NewAgent(config, []collector.Collector{mc})

	mc.Collect()

	metrics := mc.GetMetrics()
	if metrics.Counters["PollCount"] != 1 {
		t.Errorf("expected PollCount 1, got %d", metrics.Counters["PollCount"])
	}
}

func TestAgent_SendMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := Config{
		PollInterval:   1 * time.Second,
		ReportInterval: 2 * time.Second,
		ServerAddress:  server.URL,
		Logger:         zap.NewNop(),
		RateLimit:      2,
	}

	mc := newMockCollector()
	a := NewAgent(config, []collector.Collector{mc})

	metrics := mc.GetMetrics()

	err := a.sender.SendAllMetrics(metrics)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgent_Run(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := Config{
		PollInterval:   100 * time.Millisecond,
		ReportInterval: 200 * time.Millisecond,
		ServerAddress:  server.URL,
		Logger:         zap.NewNop(),
		RateLimit:      2,
	}

	mc := newMockCollector()
	a := NewAgent(config, []collector.Collector{mc})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		a.Run(ctx)
		close(done)
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()
	<-done

	metrics := mc.GetMetrics()
	if metrics.Counters["PollCount"] < 3 {
		t.Errorf("expected PollCount >= 3, got %d", metrics.Counters["PollCount"])
	}
}

func TestAgent_WorkerPool(t *testing.T) {
	var concurrent int32
	var maxConcurrent int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt32(&concurrent, 1)
		for {
			old := atomic.LoadInt32(&maxConcurrent)
			if cur <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, cur) {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&concurrent, -1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := Config{
		PollInterval:   50 * time.Millisecond,
		ReportInterval: 100 * time.Millisecond,
		ServerAddress:  server.URL,
		Logger:         zap.NewNop(),
		RateLimit:      2,
	}

	mc := newMockCollector()
	a := NewAgent(config, []collector.Collector{mc})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		a.Run(ctx)
		close(done)
	}()

	time.Sleep(600 * time.Millisecond)
	cancel()
	<-done

	if got := atomic.LoadInt32(&maxConcurrent); got > 2 {
		t.Errorf("expected max concurrent <= 2, got %d", got)
	}
}

func TestAgent_GracefulShutdown(t *testing.T) {
	var sendCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&sendCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := Config{
		PollInterval:   100 * time.Millisecond,
		ReportInterval: 200 * time.Millisecond,
		ServerAddress:  server.URL,
		Logger:         zap.NewNop(),
		RateLimit:      2,
	}

	mc := newMockCollector()
	a := NewAgent(config, []collector.Collector{mc})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		a.Run(ctx)
		close(done)
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()
	<-done

	if got := atomic.LoadInt32(&sendCount); got == 0 {
		t.Error("expected at least one metric send before shutdown")
	}
}
