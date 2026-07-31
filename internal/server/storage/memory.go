package storage

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"go.uber.org/zap"

	serverlogger "github.com/shukalov/go-ya/internal/server/logger"
	"github.com/shukalov/go-ya/pkg/models"
)

type persistStorage interface {
	Save(metrics models.RuntimeMetrics) error
	Load(metrics *models.RuntimeMetrics) error
}

type MemStorage struct {
	mu             sync.RWMutex
	gauges         map[string]float64
	counters       map[string]int64
	storeInterval  time.Duration
	persistStorage persistStorage
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func NewFileStorage(filePath string, storeInterval time.Duration) *MemStorage {
	s := NewMemStorage()
	s.storeInterval = storeInterval
	s.persistStorage = &filePersist{filePath: filePath}
	return s
}

func NewDBStorage(db *sql.DB, storeInterval time.Duration) *MemStorage {
	s := NewMemStorage()
	s.storeInterval = storeInterval
	s.persistStorage = &dbPersist{db: db}
	return s
}

func (s *MemStorage) Save() error {
	if s.persistStorage == nil {
		return nil
	}
	metrics, err := s.GetAllMetrics(context.Background())
	if err != nil {
		return err
	}
	return s.persistStorage.Save(metrics)
}

func (s *MemStorage) Load() error {
	if s.persistStorage == nil {
		return nil
	}
	var m models.RuntimeMetrics
	if err := s.persistStorage.Load(&m); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range m.Gauges {
		s.gauges[k] = v
	}
	for k, v := range m.Counters {
		s.counters[k] = v
	}

	return nil
}

func (s *MemStorage) Run() {
	if s.storeInterval <= 0 || s.persistStorage == nil {
		return
	}

	ticker := time.NewTicker(s.storeInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.Save(); err != nil {
			serverlogger.L.Error("periodic save failed", zap.Error(err))
		}
	}
}

func (s *MemStorage) UpdateGauge(_ context.Context, name string, value float64) error {
	s.mu.Lock()
	s.gauges[name] = value
	s.mu.Unlock()

	if s.persistStorage != nil && s.storeInterval == 0 {
		return s.Save()
	}

	return nil
}

func (s *MemStorage) UpdateCounter(_ context.Context, name string, value int64) error {
	s.mu.Lock()
	s.counters[name] += value
	s.mu.Unlock()

	if s.persistStorage != nil && s.storeInterval == 0 {
		return s.Save()
	}

	return nil
}

func (s *MemStorage) GetGauge(_ context.Context, name string) (float64, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.gauges[name]
	return value, ok, nil
}

func (s *MemStorage) GetCounter(_ context.Context, name string) (int64, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.counters[name]
	return value, ok, nil
}

func (s *MemStorage) GetAllMetrics(_ context.Context) (models.RuntimeMetrics, error) {
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

func (s *MemStorage) UpdateBatch(_ context.Context, metrics []models.Metrics) error {
	s.mu.Lock()

	for i := range metrics {
		switch metrics[i].MType {
		case "gauge":
			if metrics[i].Value != nil {
				s.gauges[metrics[i].ID] = *metrics[i].Value
			}
		case "counter":
			if metrics[i].Delta != nil {
				s.counters[metrics[i].ID] += *metrics[i].Delta
			}
		}
	}

	s.mu.Unlock()

	if s.persistStorage != nil && s.storeInterval == 0 {
		return s.Save()
	}

	return nil
}
