package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shukalov/go-ya/internal/server/storage"
)

func setupTestHandler() (*MetricsHandler, *storage.MemStorage) {
	storage := storage.NewMemStorage()
	handler := NewMetricsHandler(storage)
	return handler, storage
}

func TestMetricsHandler_Update_Gauge(t *testing.T) {
	handler, _ := setupTestHandler()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "valid gauge update",
			path:           "/update/gauge/test_gauge/123.45",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid gauge update with integer",
			path:           "/update/gauge/test_gauge_int/123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid gauge update with negative",
			path:           "/update/gauge/test_gauge_neg/-45.67",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBuffer(nil))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			handler.Update(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMetricsHandler_Update_Counter(t *testing.T) {
	handler, storage := setupTestHandler()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedValue  int64
	}{
		{
			name:           "valid counter update",
			path:           "/update/counter/test_counter_1/5",
			expectedStatus: http.StatusOK,
			expectedValue:  5,
		},
		{
			name:           "valid counter update with negative",
			path:           "/update/counter/test_counter_2/-10",
			expectedStatus: http.StatusOK,
			expectedValue:  -10,
		},
		{
			name:           "counter accumulates values",
			path:           "/update/counter/test_counter_3/3",
			expectedStatus: http.StatusOK,
			expectedValue:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBuffer(nil))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			handler.Update(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				// Извлекаем имя метрики из пути
				metricName := ""
				switch tt.path {
				case "/update/counter/test_counter_1/5":
					metricName = "test_counter_1"
				case "/update/counter/test_counter_2/-10":
					metricName = "test_counter_2"
				case "/update/counter/test_counter_3/3":
					metricName = "test_counter_3"
				}

				val, ok, _ := storage.GetCounter(metricName)
				if !ok {
					t.Error("expected metric to exist")
				}
				if val != tt.expectedValue {
					t.Errorf("expected %d, got %d", tt.expectedValue, val)
				}
			}
		})
	}
}

func TestMetricsHandler_Update_Counter_Accumulates(t *testing.T) {
	handler, storage := setupTestHandler()

	// Первое обновление
	req := httptest.NewRequest(http.MethodPost, "/update/counter/test_counter_acc/5", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	// Второе обновление
	req = httptest.NewRequest(http.MethodPost, "/update/counter/test_counter_acc/3", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "text/plain")
	w = httptest.NewRecorder()
	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	// Проверяем, что значения суммировались
	val, ok, _ := storage.GetCounter("test_counter_acc")
	if !ok {
		t.Error("expected metric to exist")
	}
	if val != 8 {
		t.Errorf("expected 8 (5+3), got %d", val)
	}
}

func TestMetricsHandler_Update_Gauge_Overwrites(t *testing.T) {
	handler, storage := setupTestHandler()

	// Первое обновление
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_gauge_overwrite/10.5", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	// Второе обновление с другим значением
	req = httptest.NewRequest(http.MethodPost, "/update/gauge/test_gauge_overwrite/20.7", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "text/plain")
	w = httptest.NewRecorder()
	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	// Проверяем, что значение перезаписалось
	val, ok, _ := storage.GetGauge("test_gauge_overwrite")
	if !ok {
		t.Error("expected metric to exist")
	}
	if val != 20.7 {
		t.Errorf("expected 20.7, got %f", val)
	}
}

func TestMetricsHandler_Update_Errors(t *testing.T) {
	handler, _ := setupTestHandler()

	tests := []struct {
		name           string
		method         string
		path           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "wrong method",
			method:         http.MethodGet,
			path:           "/update/gauge/test/123",
			contentType:    "text/plain",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "wrong content type",
			method:         http.MethodPost,
			path:           "/update/gauge/test/123",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing value",
			method:         http.MethodPost,
			path:           "/update/gauge/test",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing name",
			method:         http.MethodPost,
			path:           "/update/gauge",
			contentType:    "text/plain",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid gauge value",
			method:         http.MethodPost,
			path:           "/update/gauge/test/abc",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid counter value",
			method:         http.MethodPost,
			path:           "/update/counter/test/abc",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid metric type",
			method:         http.MethodPost,
			path:           "/update/invalid/test/123",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty metric name",
			method:         http.MethodPost,
			path:           "/update/gauge//123",
			contentType:    "text/plain",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(nil))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			handler.Update(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMetricsHandler_Get(t *testing.T) {
	handler, storage := setupTestHandler()

	// Добавляем тестовые метрики
	storage.UpdateGauge("test_gauge", 123.45)
	storage.UpdateCounter("test_counter", 42)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "get gauge",
			path:           "/value/gauge/test_gauge",
			expectedStatus: http.StatusOK,
			expectedBody:   "123.450000",
		},
		{
			name:           "get counter",
			path:           "/value/counter/test_counter",
			expectedStatus: http.StatusOK,
			expectedBody:   "42",
		},
		{
			name:           "not found gauge",
			path:           "/value/gauge/non_existent",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:           "invalid path",
			path:           "/value/invalid",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:           "invalid type",
			path:           "/value/invalid/test",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "empty name",
			path:           "/value/gauge/",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.Get(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" && w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestMetricsHandler_Get_NotFound(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
