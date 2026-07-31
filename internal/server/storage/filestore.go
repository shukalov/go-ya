package storage

import (
	"encoding/json"
	"os"

	"github.com/shukalov/go-ya/pkg/models"
)

type filePersist struct {
	filePath string
}

func (p *filePersist) Save(metrics models.RuntimeMetrics) error {
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

	return os.WriteFile(p.filePath, body, 0644)
}

func (p *filePersist) Load(metrics *models.RuntimeMetrics) error {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var m []models.Metrics
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	metrics.Gauges = make(map[string]float64)
	metrics.Counters = make(map[string]int64)

	for _, item := range m {
		switch item.MType {
		case "gauge":
			if item.Value != nil {
				metrics.Gauges[item.ID] = *item.Value
			}
		case "counter":
			if item.Delta != nil {
				metrics.Counters[item.ID] = *item.Delta
			}
		}
	}

	return nil
}
