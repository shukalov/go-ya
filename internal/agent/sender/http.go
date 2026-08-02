package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shukalov/go-ya/internal/agent/sender/httperrors"
	"github.com/shukalov/go-ya/internal/retry"
	"github.com/shukalov/go-ya/pkg/models"
)

type HTTPSender struct {
	serverAddress string
	client        *http.Client
}

func NewHTTPSender(serverAddress string) *HTTPSender {
	return &HTTPSender{
		serverAddress: serverAddress,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *HTTPSender) sendJSON(path string, data []byte) error {
	retryCfg := retry.Config{
		StrategyCfg: retry.DefaultConfig.StrategyCfg,
		Classify:  httperrors.NewHTTPErrorClassifier().Classify,
	}

	return retry.Execute(context.Background(), retryCfg, func() error {
		var buf bytes.Buffer
		gz, err := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
		if err != nil {
			return fmt.Errorf("failed to create gzip writer: %w", err)
		}
		if _, err := gz.Write(data); err != nil {
			return fmt.Errorf("failed to compress body: %w", err)
		}
		if err := gz.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}

		req, err := http.NewRequest(http.MethodPost, s.serverAddress+path, &buf)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := s.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		return nil
	})
}

func (s *HTTPSender) SendMetric(metricType string, name string, value interface{}) error {
	m := models.Metrics{
		ID:    name,
		MType: metricType,
	}

	switch v := value.(type) {
	case float64:
		m.Value = &v
	case int64:
		m.Delta = &v
	default:
		return fmt.Errorf("unsupported value type: %T", value)
	}

	body, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	return s.sendJSON("/update", body)
}

func (s *HTTPSender) SendAllMetrics(metrics models.RuntimeMetrics) error {
	var batch []models.Metrics

	for name := range models.GaugeDefs() {
		v := metrics.Gauges[name]
		batch = append(batch, models.Metrics{ID: name, MType: "gauge", Value: &v})
	}

	for name := range models.CounterDefs() {
		v := metrics.Counters[name]
		batch = append(batch, models.Metrics{ID: name, MType: "counter", Delta: &v})
	}

	return s.SendBatch(batch)
}

func (s *HTTPSender) SendBatch(metrics []models.Metrics) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return s.sendJSON("/updates", body)
}
