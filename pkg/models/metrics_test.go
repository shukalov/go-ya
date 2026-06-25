package models

import (
	"testing"
)

func TestMetricTypes(t *testing.T) {
	tests := []struct {
		name       string
		metricType MetricType
		expected   string
	}{
		{
			name:       "gauge type",
			metricType: TypeGauge,
			expected:   "gauge",
		},
		{
			name:       "counter type",
			metricType: TypeCounter,
			expected:   "counter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.metricType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.metricType)
			}
		})
	}
}

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
