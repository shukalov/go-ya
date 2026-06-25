package sender

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
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
func (s *HTTPSender) SendMetric(metricType models.MetricType, name string, value interface{}) error {
	var valueStr string
	switch v := value.(type) {
	case float64:
		valueStr = strconv.FormatFloat(v, 'f', -1, 64)
	case int64:
		valueStr = strconv.FormatInt(v, 10)
	default:
		return fmt.Errorf("unsupported value type: %T", value)
	}

	url := fmt.Sprintf("%s/update/%s/%s/%s",
		s.serverAddress,
		metricType,
		name,
		valueStr)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

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
	for _, def := range models.GaugeDefs() {
		if err := s.SendMetric(models.TypeGauge, def.Name, metrics.Gauges[def.Name]); err != nil {
			return fmt.Errorf("failed to send gauge %s: %w", def.Name, err)
		}
	}

	for _, def := range models.CounterDefs() {
		if err := s.SendMetric(models.TypeCounter, def.Name, metrics.Counters[def.Name]); err != nil {
			return fmt.Errorf("failed to send counter %s: %w", def.Name, err)
		}
	}

	return nil
}
