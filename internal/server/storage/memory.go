package storage

import (
	"sync"

	"github.com/shukalov/go-ya/pkg/models"
)

// MemStorage - in-memory хранилище метрик
type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

// NewMemStorage - создает новое хранилище
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// UpdateGauge - обновляет или добавляет gauge метрику
func (s *MemStorage) UpdateGauge(name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value
	return nil
}

// UpdateCounter - обновляет или добавляет counter метрику
func (s *MemStorage) UpdateCounter(name string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += value
	return nil
}

// GetGauge - получает значение gauge метрики
func (s *MemStorage) GetGauge(name string) (float64, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.gauges[name]
	return value, ok, nil
}

// GetCounter - получает значение counter метрики
func (s *MemStorage) GetCounter(name string) (int64, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.counters[name]
	return value, ok, nil
}

// GetAllMetrics - возвращает все метрики
func (s *MemStorage) GetAllMetrics() (models.RuntimeMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gauges := make(map[string]float64)
	for k, v := range s.gauges {
		gauges[k] = v
	}

	counters := make(map[string]int64)
	for k, v := range s.counters {
		counters[k] = v
	}

	return models.RuntimeMetrics{Gauges: gauges, Counters: counters}, nil
}
