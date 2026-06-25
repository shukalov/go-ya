package storage

// Storage - интерфейс для работы с хранилищем метрик
type Storage interface {
	// UpdateGauge обновляет или добавляет gauge метрику
	UpdateGauge(name string, value float64) error

	// UpdateCounter обновляет или добавляет counter метрику
	UpdateCounter(name string, value int64) error

	// GetGauge возвращает значение gauge метрики
	GetGauge(name string) (float64, bool, error)

	// GetCounter возвращает значение counter метрики
	GetCounter(name string) (int64, bool, error)

	// GetAllMetrics возвращает все метрики (для отладки)
	GetAllMetrics() (map[string]float64, map[string]int64, error)
}
