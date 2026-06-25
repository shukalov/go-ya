package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/shukalov/go-ya/internal/server/storage"
)

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
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 || parts[0] != "update" {
		http.Error(w, "Invalid endpoint", http.StatusNotFound)
		return
	}

	if len(parts) < 4 {
		if len(parts) >= 3 {
			http.Error(w, "Value is required", http.StatusBadRequest)
		} else {
			http.Error(w, "Metric name is required", http.StatusNotFound)
		}
		return
	}

	metricType := parts[1]
	metricName := parts[2]
	valueStr := parts[3]

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	if !isValidMetricName(metricName) {
		http.Error(w, fmt.Sprintf("Invalid metric name: '%s'", metricName), http.StatusBadRequest)
		return
	}

	if valueStr == "" {
		http.Error(w, "Value is required", http.StatusBadRequest)
		return
	}

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid gauge value: '%s'", valueStr), http.StatusBadRequest)
			return
		}
		if err := h.storage.UpdateGauge(metricName, value); err != nil {
			http.Error(w, fmt.Sprintf("Storage error: %v", err), http.StatusInternalServerError)
			return
		}

	case "counter":
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid counter value: '%s'", valueStr), http.StatusBadRequest)
			return
		}
		if err := h.storage.UpdateCounter(metricName, value); err != nil {
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
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 || parts[0] != "value" {
		http.Error(w, "Invalid path. Expected: /value/{type}/{name}", http.StatusNotFound)
		return
	}

	metricType := parts[1]
	metricName := parts[2]

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	switch metricType {
	case "gauge":
		value, ok, err := h.storage.GetGauge(metricName)
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
		value, ok, err := h.storage.GetCounter(metricName)
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
