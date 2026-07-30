package storage

import (
	"context"

	"github.com/shukalov/go-ya/pkg/models"
)

// Storage - интерфейс для работы с хранилищем метрик
type Storage interface {
	// UpdateGauge обновляет или добавляет gauge метрику
	UpdateGauge(ctx context.Context, name string, value float64) error

	// UpdateCounter обновляет или добавляет counter метрику
	UpdateCounter(ctx context.Context, name string, value int64) error

	// GetGauge возвращает значение gauge метрики
	GetGauge(ctx context.Context, name string) (float64, bool, error)

	// GetCounter возвращает значение counter метрики
	GetCounter(ctx context.Context, name string) (int64, bool, error)

	// GetAllMetrics возвращает все метрики
	GetAllMetrics(ctx context.Context) (models.RuntimeMetrics, error)
}
