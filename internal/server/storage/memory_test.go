package storage

import (
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	if storage == nil {
		t.Error("expected storage to be created")
	}
	if storage.gauges == nil {
		t.Error("expected gauges map to be initialized")
	}
	if storage.counters == nil {
		t.Error("expected counters map to be initialized")
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	storage := NewMemStorage()

	tests := []struct {
		name  string
		key   string
		value float64
	}{
		{
			name:  "update gauge",
			key:   "test_gauge",
			value: 123.45,
		},
		{
			name:  "update gauge with zero",
			key:   "test_gauge_zero",
			value: 0.0,
		},
		{
			name:  "update gauge with negative",
			key:   "test_gauge_negative",
			value: -123.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.UpdateGauge(tt.key, tt.value)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			val, ok, err := storage.GetGauge(tt.key)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !ok {
				t.Error("expected metric to exist")
			}
			if val != tt.value {
				t.Errorf("expected %f, got %f", tt.value, val)
			}
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	storage := NewMemStorage()

	tests := []struct {
		name     string
		key      string
		values   []int64
		expected int64
	}{
		{
			name:     "single increment",
			key:      "test_counter_1",
			values:   []int64{5},
			expected: 5,
		},
		{
			name:     "multiple increments",
			key:      "test_counter_2",
			values:   []int64{1, 2, 3, 4, 5},
			expected: 15,
		},
		{
			name:     "with negative values",
			key:      "test_counter_3",
			values:   []int64{10, -5, 3},
			expected: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, v := range tt.values {
				err := storage.UpdateCounter(tt.key, v)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			val, ok, err := storage.GetCounter(tt.key)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !ok {
				t.Error("expected metric to exist")
			}
			if val != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, val)
			}
		})
	}
}

func TestMemStorage_GetGauge_NotFound(t *testing.T) {
	storage := NewMemStorage()
	val, ok, err := storage.GetGauge("non_existent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected metric not to exist")
	}
	if val != 0 {
		t.Errorf("expected 0, got %f", val)
	}
}

func TestMemStorage_GetCounter_NotFound(t *testing.T) {
	storage := NewMemStorage()
	val, ok, err := storage.GetCounter("non_existent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected metric not to exist")
	}
	if val != 0 {
		t.Errorf("expected 0, got %d", val)
	}
}

func TestMemStorage_GetAllMetrics(t *testing.T) {
	storage := NewMemStorage()

	// Добавляем тестовые метрики
	storage.UpdateGauge("gauge1", 1.1)
	storage.UpdateGauge("gauge2", 2.2)
	storage.UpdateCounter("counter1", 10)
	storage.UpdateCounter("counter2", 20)

	gauges, counters, err := storage.GetAllMetrics()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(gauges) != 2 {
		t.Errorf("expected 2 gauges, got %d", len(gauges))
	}
	if len(counters) != 2 {
		t.Errorf("expected 2 counters, got %d", len(counters))
	}

	if gauges["gauge1"] != 1.1 {
		t.Errorf("expected 1.1, got %f", gauges["gauge1"])
	}
	if counters["counter1"] != 10 {
		t.Errorf("expected 10, got %d", counters["counter1"])
	}
}

func TestMemStorage_Concurrency(t *testing.T) {
	storage := NewMemStorage()
	done := make(chan bool)

	// Запускаем несколько горутин для теста конкурентности
	for i := 0; i < 100; i++ {
		go func(id int) {
			storage.UpdateGauge("gauge", float64(id))
			storage.UpdateCounter("counter", int64(id))
			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < 100; i++ {
		<-done
	}

	val, ok, _ := storage.GetGauge("gauge")
	if !ok {
		t.Error("expected gauge to exist")
	}
	// Значение должно быть последним записанным
	if val < 0 || val > 100 {
		t.Errorf("unexpected gauge value: %f", val)
	}
}
