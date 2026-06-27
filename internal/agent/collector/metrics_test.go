package collector

import (
	"testing"
)

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	if collector == nil {
		t.Error("expected collector to be created")
	}
	if collector.metrics.Gauges == nil {
		t.Error("expected gauges map to be initialized")
	}
	if collector.metrics.Counters["PollCount"] != 0 {
		t.Errorf("expected PollCount 0, got %d", collector.metrics.Counters["PollCount"])
	}
}

func TestMetricsCollector_Collect(t *testing.T) {
	collector := NewMetricsCollector()

	// Собираем метрики
	collector.Collect()

	// Проверяем, что метрики заполнены
	metrics := collector.GetMetrics()

	// Проверяем, что PollCount увеличился
	if metrics.Counters["PollCount"] != 1 {
		t.Errorf("expected PollCount 1, got %d", metrics.Counters["PollCount"])
	}

	// Проверяем, что RandomValue в допустимом диапазоне
	if metrics.Gauges["RandomValue"] < 0 || metrics.Gauges["RandomValue"] > 100 {
		t.Errorf("RandomValue out of range: %f", metrics.Gauges["RandomValue"])
	}

	// Проверяем несколько метрик из runtime
	if metrics.Gauges["Alloc"] < 0 {
		t.Errorf("Alloc should be >= 0, got %f", metrics.Gauges["Alloc"])
	}
	if metrics.Gauges["Sys"] < 0 {
		t.Errorf("Sys should be >= 0, got %f", metrics.Gauges["Sys"])
	}
}

func TestMetricsCollector_MultipleCollect(t *testing.T) {
	collector := NewMetricsCollector()

	// Собираем метрики несколько раз
	for i := int64(1); i <= 5; i++ {
		collector.Collect()
		metrics := collector.GetMetrics()
		if metrics.Counters["PollCount"] != i {
			t.Errorf("expected PollCount %d, got %d", i, metrics.Counters["PollCount"])
		}
	}
}

func TestMetricsCollector_Concurrency(t *testing.T) {
	collector := NewMetricsCollector()
	done := make(chan bool)

	// Запускаем несколько горутин для сбора метрик
	for i := 0; i < 10; i++ {
		go func() {
			collector.Collect()
			done <- true
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}

	// Проверяем, что счетчик увеличился корректно
	metrics := collector.GetMetrics()
	if metrics.Counters["PollCount"] != 10 {
		t.Errorf("expected PollCount 10, got %d", metrics.Counters["PollCount"])
	}
}

func TestMetricsCollector_GetMetrics_Copy(t *testing.T) {
	collector := NewMetricsCollector()
	collector.Collect()

	// Получаем копию метрик
	metrics1 := collector.GetMetrics()
	metrics2 := collector.GetMetrics()

	// Изменяем первую копию
	metrics1.Gauges["Alloc"] = 999.9

	// Проверяем, что вторая копия не изменилась
	if metrics2.Gauges["Alloc"] == 999.9 {
		t.Error("expected metrics to be copied, not referenced")
	}
}
