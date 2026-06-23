package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// MemStorage - структура для хранения метрик
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

// Storage - интерфейс для работы с хранилищем
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}

// UpdateGauge - обновляет или добавляет gauge метрику
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value
}

// UpdateCounter - обновляет или добавляет counter метрику
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += value
}

// GetGauge - получает значение gauge метрики
func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.gauges[name]
	return value, ok
}

// GetCounter - получает значение counter метрики
func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.counters[name]
	return value, ok
}

// isValidMetricName - проверяет, является ли имя метрики валидным
// Имя может содержать буквы, цифры, подчеркивания и точки
func isValidMetricName(name string) bool {
	if name == "" {
		return false
	}
	// Разрешаем буквы (включая Unicode), цифры, подчеркивания и точки
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '.') {
			return false
		}
	}
	return true
}

// updateHandler - обработчик для обновления метрик
func updateHandler(storage *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод запроса
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

		// Проверяем, что путь начинается с "update"
		if len(parts) < 1 || parts[0] != "update" {
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
			return
		}

		// Проверяем наличие всех частей URL
		// Ожидаем: /update/{type}/{name}/{value}
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

		// Проверяем наличие имени метрики
		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		// Проверяем, что имя метрики содержит только допустимые символы
		if !isValidMetricName(metricName) {
			http.Error(w, fmt.Sprintf("Invalid metric name: '%s' contains invalid characters", metricName), http.StatusBadRequest)
			return
		}

		// Проверяем наличие значения
		if valueStr == "" {
			http.Error(w, "Value is required", http.StatusBadRequest)
			return
		}

		// Обрабатываем в зависимости от типа метрики
		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid gauge value: '%s' is not a valid number", valueStr), http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(metricName, value)

		case "counter":
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid counter value: '%s' is not a valid integer", valueStr), http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(metricName, value)

		default:
			http.Error(w, fmt.Sprintf("Invalid metric type: '%s'. Allowed: gauge, counter", metricType), http.StatusBadRequest)
			return
		}

		// Успешный ответ
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	}
}

// getHandler - обработчик для получения значений метрик (для тестирования)
func getHandler(storage *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")

		// Ожидаем /value/{type}/{name}
		if len(parts) != 3 || parts[0] != "value" {
			http.Error(w, "Invalid path. Expected: /value/{type}/{name}", http.StatusNotFound)
			return
		}

		metricType := parts[1]
		metricName := parts[2]

		// Проверяем наличие имени метрики
		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		switch metricType {
		case "gauge":
			value, ok := storage.GetGauge(metricName)
			if !ok {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%f", value)

		case "counter":
			value, ok := storage.GetCounter(metricName)
			if !ok {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%d", value)

		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
	}
}

func main() {
	// Создаем хранилище
	storage := NewMemStorage()

	// Регистрируем обработчики
	http.HandleFunc("/update/", updateHandler(storage))
	http.HandleFunc("/value/", getHandler(storage))

	fmt.Println("Server is running on http://localhost:8080")

	// Запускаем сервер
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
