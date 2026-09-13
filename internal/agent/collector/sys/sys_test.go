package sys

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	c := New(2*time.Second, zap.NewNop())
	if c == nil {
		t.Error("expected collector to be created")
	}
}

func TestSysCollector_Collect(t *testing.T) {
	c := New(2*time.Second, zap.NewNop())
	c.Collect()

	metrics := c.GetMetrics()

	if metrics.Gauges["TotalMemory"] <= 0 {
		t.Errorf("TotalMemory should be > 0, got %f", metrics.Gauges["TotalMemory"])
	}
	if metrics.Gauges["FreeMemory"] < 0 {
		t.Errorf("FreeMemory should be >= 0, got %f", metrics.Gauges["FreeMemory"])
	}
}

func TestSysCollector_CPUutilization(t *testing.T) {
	c := New(2*time.Second, zap.NewNop())
	c.Collect()

	metrics := c.GetMetrics()
	numCPU := runtime.NumCPU()

	if numCPU <= 0 {
		t.Skip("no CPUs detected")
	}

	for i := 0; i < numCPU; i++ {
		key := fmt.Sprintf("CPUutilization1_%d", i)
		v, ok := metrics.Gauges[key]
		if !ok {
			t.Errorf("expected gauge %s", key)
		}
		if v < 0 || v > 100 {
			t.Errorf("%s out of range: %f", key, v)
		}
	}
}

func TestSysCollector_GetMetrics_Copy(t *testing.T) {
	c := New(2*time.Second, zap.NewNop())
	c.Collect()

	m1 := c.GetMetrics()
	m2 := c.GetMetrics()

	m1.Gauges["TotalMemory"] = 999

	if m2.Gauges["TotalMemory"] == 999 {
		t.Error("expected metrics to be copied, not referenced")
	}
}
