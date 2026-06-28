package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shukalov/go-ya/pkg/models"
)

// HTTPSender - отправляет метрики по HTTP
type HTTPSender struct {
	serverAddress string
	client        *http.Client
}

// NewHTTPSender - создает новый HTTP отправитель
func NewHTTPSender(serverAddress string) *HTTPSender {
	return &HTTPSender{
		serverAddress: serverAddress,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendMetric - отправляет одну метрику
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

	var buf bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	if _, err := gz.Write(body); err != nil {
		return fmt.Errorf("failed to compress body: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.serverAddress+"/update", &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// SendAllMetrics - отправляет все метрики
func (s *HTTPSender) SendAllMetrics(metrics models.RuntimeMetrics) error {
	for name := range models.GaugeDefs() {
		if err := s.SendMetric("gauge", name, metrics.Gauges[name]); err != nil {
			return fmt.Errorf("failed to send gauge %s: %w", name, err)
		}
	}

	for name := range models.CounterDefs() {
		if err := s.SendMetric("counter", name, metrics.Counters[name]); err != nil {
			return fmt.Errorf("failed to send counter %s: %w", name, err)
		}
	}

	return nil
}
