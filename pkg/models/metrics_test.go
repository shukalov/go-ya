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
	metrics := &RuntimeMetrics{
		Alloc:       1024.5,
		PollCount:   5,
		RandomValue: 42.0,
	}

	if metrics.Alloc != 1024.5 {
		t.Errorf("expected Alloc 1024.5, got %f", metrics.Alloc)
	}
	if metrics.PollCount != 5 {
		t.Errorf("expected PollCount 5, got %d", metrics.PollCount)
	}
	if metrics.RandomValue != 42.0 {
		t.Errorf("expected RandomValue 42.0, got %f", metrics.RandomValue)
	}
}
