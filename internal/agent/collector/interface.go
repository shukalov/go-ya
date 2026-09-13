package collector

import (
	"context"

	"github.com/shukalov/go-ya/pkg/models"
)

// Collector - интерфейс сборщика метрик
type Collector interface {
	Run(ctx context.Context)
	GetMetrics() models.RuntimeMetrics
}
