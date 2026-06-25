package collector

import (
	"testing"
)

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	if collector == nil {
		t.Error("expected collector to be created")
	}
	if collector.metrics == nil {
		t.Error("expected metrics to be initialized")
	}
	if collector.pollCount != 0 {
		t.Errorf("expected pollCount 0, got %d", collector.pollCount)
	}
}

func TestMetricsCollector_Collect(t *testing.T) {
	collector := NewMetricsCollector()

	// Собираем метрики
	collector.Collect()

	// Проверяем, что метрики заполнены
	metrics := collector.GetMetrics()

	// Проверяем, что PollCount увеличился
	if metrics.PollCount != 1 {
		t.Errorf("expected PollCount 1, got %d", metrics.PollCount)
	}

	// Проверяем, что RandomValue в допустимом диапазоне
	if metrics.RandomValue < 0 || metrics.RandomValue > 100 {
		t.Errorf("RandomValue out of range: %f", metrics.RandomValue)
	}

	// Проверяем несколько метрик из runtime
	if metrics.Alloc < 0 {
		t.Errorf("Alloc should be >= 0, got %f", metrics.Alloc)
	}
	if metrics.Sys < 0 {
		t.Errorf("Sys should be >= 0, got %f", metrics.Sys)
	}
}

func TestMetricsCollector_MultipleCollect(t *testing.T) {
	collector := NewMetricsCollector()

	// Собираем метрики несколько раз
	for i := 1; i <= 5; i++ {
		collector.Collect()
		pollCount := collector.GetPollCount()
		if pollCount != int64(i) {
			t.Errorf("expected PollCount %d, got %d", i, pollCount)
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
	pollCount := collector.GetPollCount()
	if pollCount != 10 {
		t.Errorf("expected PollCount 10, got %d", pollCount)
	}
}

func TestMetricsCollector_GetMetrics_Copy(t *testing.T) {
	collector := NewMetricsCollector()
	collector.Collect()

	// Получаем копию метрик
	metrics1 := collector.GetMetrics()
	metrics2 := collector.GetMetrics()

	// Изменяем первую копию
	metrics1.Alloc = 999.9

	// Проверяем, что вторая копия не изменилась
	if metrics2.Alloc == 999.9 {
		t.Error("expected metrics to be copied, not referenced")
	}
}
