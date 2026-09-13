package runtime

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	c := New(2*time.Second, zap.NewNop())
	if c == nil {
		t.Error("expected collector to be created")
	}
	if c.metrics.Gauges == nil {
		t.Error("expected gauges map to be initialized")
	}
	if c.metrics.Counters["PollCount"] != 0 {
		t.Errorf("expected PollCount 0, got %d", c.metrics.Counters["PollCount"])
	}
}

func TestRuntimeCollector_Collect(t *testing.T) {
	c := New(2*time.Second, zap.NewNop())
	c.Collect()

	metrics := c.GetMetrics()

	if metrics.Counters["PollCount"] != 1 {
		t.Errorf("expected PollCount 1, got %d", metrics.Counters["PollCount"])
	}

	if metrics.Gauges["RandomValue"] < 0 || metrics.Gauges["RandomValue"] > 100 {
		t.Errorf("RandomValue out of range: %f", metrics.Gauges["RandomValue"])
	}
	if metrics.Gauges["Alloc"] < 0 {
		t.Errorf("Alloc should be >= 0, got %f", metrics.Gauges["Alloc"])
	}
	if metrics.Gauges["Sys"] < 0 {
		t.Errorf("Sys should be >= 0, got %f", metrics.Gauges["Sys"])
	}
}

func TestRuntimeCollector_MultipleCollect(t *testing.T) {
	c := New(2*time.Second, zap.NewNop())

	for i := int64(1); i <= 5; i++ {
		c.Collect()
		metrics := c.GetMetrics()
		if metrics.Counters["PollCount"] != i {
			t.Errorf("expected PollCount %d, got %d", i, metrics.Counters["PollCount"])
		}
	}
}

func TestRuntimeCollector_Concurrency(t *testing.T) {
	c := New(2*time.Second, zap.NewNop())
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			c.Collect()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := c.GetMetrics()
	if metrics.Counters["PollCount"] != 10 {
		t.Errorf("expected PollCount 10, got %d", metrics.Counters["PollCount"])
	}
}

func TestRuntimeCollector_GetMetrics_Copy(t *testing.T) {
	c := New(2*time.Second, zap.NewNop())
	c.Collect()

	metrics1 := c.GetMetrics()
	metrics2 := c.GetMetrics()

	metrics1.Gauges["Alloc"] = 999.9

	if metrics2.Gauges["Alloc"] == 999.9 {
		t.Error("expected metrics to be copied, not referenced")
	}
}
