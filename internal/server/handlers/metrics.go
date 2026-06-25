package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/shukalov/go-ya/internal/server/storage"
)

type parsedPath struct {
	endpoint   string
	metricType string
	metricName string
}

func parsePath(path string) (parsedPath, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return parsedPath{}, fmt.Errorf("invalid path")
	}
	pp := parsedPath{
		endpoint:   parts[0],
		metricType: parts[1],
		metricName: parts[2],
	}
	if !isValidMetricName(pp.metricName) {
		return parsedPath{}, fmt.Errorf("invalid metric name: '%s'", pp.metricName)
	}
	return pp, nil
}

// MetricsHandler - обработчик для метрик
type MetricsHandler struct {
	storage storage.Storage
}

// NewMetricsHandler - создает новый обработчик
func NewMetricsHandler(storage storage.Storage) *MetricsHandler {
	return &MetricsHandler{
		storage: storage,
	}
}

// isValidMetricName - проверяет валидность имени метрики
func isValidMetricName(name string) bool {
	if name == "" {
		return false
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '.') {
			return false
		}
	}
	return true
}

// Update - обработчик для обновления метрик
func (h *MetricsHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	// Разбираем путь
	pp, err := parsePath(r.URL.Path)
	if err != nil || pp.endpoint != "update" {
		http.Error(w, "Invalid endpoint", http.StatusNotFound)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[3] == "" {
		http.Error(w, "Value is required", http.StatusBadRequest)
		return
	}
	valueStr := parts[3]

	switch pp.metricType {
	case "gauge":
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid gauge value: '%s'", valueStr), http.StatusBadRequest)
			return
		}
		if err := h.storage.UpdateGauge(pp.metricName, value); err != nil {
			http.Error(w, fmt.Sprintf("Storage error: %v", err), http.StatusInternalServerError)
			return
		}

	case "counter":
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid counter value: '%s'", valueStr), http.StatusBadRequest)
			return
		}
		if err := h.storage.UpdateCounter(pp.metricName, value); err != nil {
			http.Error(w, fmt.Sprintf("Storage error: %v", err), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Invalid metric type. Allowed: gauge, counter", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// Get - обработчик для получения значений метрик
func (h *MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {
	pp, err := parsePath(r.URL.Path)
	if err != nil || pp.endpoint != "value" {
		http.Error(w, "Invalid path. Expected: /value/{type}/{name}", http.StatusNotFound)
		return
	}

	switch pp.metricType {
	case "gauge":
		value, ok, err := h.storage.GetGauge(pp.metricName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Storage error: %v", err), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%f", value)

	case "counter":
		value, ok, err := h.storage.GetCounter(pp.metricName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Storage error: %v", err), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%d", value)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
	}
}
