package models

import (
	"testing"
)

func TestRuntimeMetrics(t *testing.T) {
	metrics := RuntimeMetrics{
		Gauges:   map[string]float64{"Alloc": 1024.5, "RandomValue": 42.0},
		Counters: map[string]int64{"PollCount": 5},
	}

	if metrics.Gauges["Alloc"] != 1024.5 {
		t.Errorf("expected Alloc 1024.5, got %f", metrics.Gauges["Alloc"])
	}
	if metrics.Counters["PollCount"] != 5 {
		t.Errorf("expected PollCount 5, got %d", metrics.Counters["PollCount"])
	}
	if metrics.Gauges["RandomValue"] != 42.0 {
		t.Errorf("expected RandomValue 42.0, got %f", metrics.Gauges["RandomValue"])
	}
}
