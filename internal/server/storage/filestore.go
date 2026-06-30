package storage

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	serverlogger "github.com/shukalov/go-ya/internal/server/logger"
	"github.com/shukalov/go-ya/pkg/models"
)

type FileBackedStorage struct {
	MemStorage
	filePath      string
	storeInterval time.Duration
	fileMu        sync.Mutex
}

func NewFileBackedStorage(filePath string, storeInterval time.Duration) *FileBackedStorage {
	return &FileBackedStorage{
		MemStorage:    *NewMemStorage(),
		filePath:      filePath,
		storeInterval: storeInterval,
	}
}

func (fs *FileBackedStorage) Save() error {
	fs.fileMu.Lock()
	defer fs.fileMu.Unlock()

	metrics, err := fs.GetAllMetrics()
	if err != nil {
		return err
	}

	var data []models.Metrics
	for name, val := range metrics.Gauges {
		v := val
		data = append(data, models.Metrics{ID: name, MType: "gauge", Value: &v})
	}
	for name, val := range metrics.Counters {
		v := val
		data = append(data, models.Metrics{ID: name, MType: "counter", Delta: &v})
	}

	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(fs.filePath, body, 0644); err != nil {
		return err
	}

	serverlogger.L.Info("saved metrics to file", zap.String("path", fs.filePath))
	return nil
}

func (fs *FileBackedStorage) Load() error {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			if m.Value != nil {
				fs.gauges[m.ID] = *m.Value
			}
		case "counter":
			if m.Delta != nil {
				fs.counters[m.ID] = *m.Delta
			}
		}
	}

	serverlogger.L.Info("loaded metrics from file", zap.String("path", fs.filePath))
	return nil
}

func (fs *FileBackedStorage) Run() {
	if fs.storeInterval <= 0 {
		return
	}

	ticker := time.NewTicker(fs.storeInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := fs.Save(); err != nil {
			serverlogger.L.Error("periodic save failed", zap.Error(err))
		}
	}
}
