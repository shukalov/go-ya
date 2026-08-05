package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shukalov/go-ya/internal/server/storage"
	"github.com/shukalov/go-ya/pkg/models"
)

func metricDescription(name string) string {
	if def, ok := models.MetricDefs()[name]; ok {
		return def.Description
	}
	return ""
}

func setupRouter() (*gin.Engine, *storage.MemStorage) {
	gin.SetMode(gin.TestMode)
	store := storage.NewMemStorage()
	handler := NewMetricsHandler(store)

	router := gin.New()
	router.POST("/update/:type/:name/:value", handler.Update)
	router.POST("/update", handler.UpdateJSON)
	router.GET("/value/:type/:name", handler.Get)
	router.POST("/value", handler.GetJSON)
	router.GET("/", handler.Index)
	router.GET("/metrics", handler.Metrics)

	return router, store
}

func TestMetricsHandler_Update_Gauge(t *testing.T) {
	router, _ := setupRouter()

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
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, tt.path, nil)
			req.Header.Set("Content-Type", "text/plain")
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMetricsHandler_Update_Counter(t *testing.T) {
	router, store := setupRouter()

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
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, tt.path, nil)
			req.Header.Set("Content-Type", "text/plain")
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				metricName := ""
				switch tt.path {
				case "/update/counter/test_counter_1/5":
					metricName = "test_counter_1"
				case "/update/counter/test_counter_2/-10":
					metricName = "test_counter_2"
				case "/update/counter/test_counter_3/3":
					metricName = "test_counter_3"
				}

				val, ok, _ := store.GetCounter(context.Background(),metricName)
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
	router, store := setupRouter()

	req, _ := http.NewRequest(http.MethodPost, "/update/counter/test_counter_acc/5", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	req, _ = http.NewRequest(http.MethodPost, "/update/counter/test_counter_acc/3", nil)
	req.Header.Set("Content-Type", "text/plain")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	val, ok, _ := store.GetCounter(context.Background(),"test_counter_acc")
	if !ok {
		t.Error("expected metric to exist")
	}
	if val != 8 {
		t.Errorf("expected 8 (5+3), got %d", val)
	}
}

func TestMetricsHandler_Update_Gauge_Overwrites(t *testing.T) {
	router, store := setupRouter()

	req, _ := http.NewRequest(http.MethodPost, "/update/gauge/test_gauge_overwrite/10.5", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	req, _ = http.NewRequest(http.MethodPost, "/update/gauge/test_gauge_overwrite/20.7", nil)
	req.Header.Set("Content-Type", "text/plain")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	val, ok, _ := store.GetGauge(context.Background(),"test_gauge_overwrite")
	if !ok {
		t.Error("expected metric to exist")
	}
	if val != 20.7 {
		t.Errorf("expected 20.7, got %f", val)
	}
}

func TestMetricsHandler_Update_Errors(t *testing.T) {
	router, _ := setupRouter()

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
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "wrong content type",
			method:         http.MethodPost,
			path:           "/update/gauge/test/123",
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing value",
			method:         http.MethodPost,
			path:           "/update/gauge/test",
			contentType:    "text/plain",
			expectedStatus: http.StatusNotFound,
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
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", tt.contentType)
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMetricsHandler_Get(t *testing.T) {
	router, store := setupRouter()

	store.UpdateGauge(context.Background(),"test_gauge", 123.45)
	store.UpdateCounter(context.Background(),"test_counter", 42)

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
			expectedBody:   "123.45",
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
			expectedStatus: http.StatusMovedPermanently,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(w, req)

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
	router, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/value/gauge/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestMetricsHandler_Index(t *testing.T) {
	router, store := setupRouter()

	store.UpdateGauge(context.Background(),"Alloc", 1024.5)
	store.UpdateCounter(context.Background(),"PollCount", 42)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html, got %s", ct)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Alloc") {
		t.Errorf("expected body to contain Alloc")
	}
	if !strings.Contains(body, "1024.5") {
		t.Errorf("expected body to contain 1024.5")
	}
	if !strings.Contains(body, "PollCount") {
		t.Errorf("expected body to contain PollCount")
	}
	if !strings.Contains(body, "42") {
		t.Errorf("expected body to contain 42")
	}
}

func TestIsValidMetricName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		valid  bool
	}{
		{name: "alphanumeric", input: "test123", valid: true},
		{name: "with underscore", input: "test_metric", valid: true},
		{name: "with dot", input: "test.metric", valid: true},
		{name: "mixed case", input: "TestMetric_1.0", valid: true},
		{name: "empty", input: "", valid: false},
		{name: "with space", input: "test metric", valid: false},
		{name: "with hyphen", input: "test-metric", valid: false},
		{name: "with slash", input: "test/metric", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidMetricName(tt.input)
			if got != tt.valid {
				t.Errorf("isValidMetricName(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestMetricsHandler_Index_Empty(t *testing.T) {
	router, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<table>") {
		t.Errorf("expected body to contain table element")
	}
	if !strings.Contains(body, "Metrics") {
		t.Errorf("expected body to contain Metrics heading")
	}
	if strings.Contains(body, "<tr><td>") {
		t.Errorf("expected no metric rows in empty state")
	}
}

func TestMetricsHandler_Index_Partial(t *testing.T) {
	router, store := setupRouter()

	store.UpdateGauge(context.Background(),"Alloc", 1024.5)
	store.UpdateGauge(context.Background(),"Sys", 512.0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Alloc") {
		t.Errorf("expected body to contain Alloc")
	}
	if !strings.Contains(body, "Sys") {
		t.Errorf("expected body to contain Sys")
	}
	if strings.Contains(body, "PollCount") {
		t.Errorf("expected no counter metrics")
	}
	if !strings.Contains(body, "gauge") {
		t.Errorf("expected type column for metrics")
	}
}

func TestMetricsHandler_Metrics_Empty(t *testing.T) {
	router, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if body != "" {
		t.Errorf("expected empty body, got %q", body)
	}
}

func TestMetricsHandler_Update_Gauge_ZeroAndLarge(t *testing.T) {
	router, store := setupRouter()

	tests := []struct {
		name           string
		path           string
		expectedValue  float64
	}{
		{name: "zero value", path: "/update/gauge/zero_gauge/0", expectedValue: 0},
		{name: "large value", path: "/update/gauge/large_gauge/1.7976931348623157e+308", expectedValue: 1.7976931348623157e+308},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, tt.path, nil)
			req.Header.Set("Content-Type", "text/plain")
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			metricName := ""
			switch tt.path {
			case "/update/gauge/zero_gauge/0":
				metricName = "zero_gauge"
			case "/update/gauge/large_gauge/1.7976931348623157e+308":
				metricName = "large_gauge"
			}

			val, ok, _ := store.GetGauge(context.Background(),metricName)
			if !ok {
				t.Error("expected metric to exist")
			}
			if val != tt.expectedValue {
				t.Errorf("expected %f, got %f", tt.expectedValue, val)
			}
		})
	}
}

func TestMetricsHandler_Metrics_Prometheus(t *testing.T) {
	router, store := setupRouter()

	store.UpdateGauge(context.Background(),"Alloc", 1024.5)
	store.UpdateCounter(context.Background(),"PollCount", 42)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected text/plain, got %s", ct)
	}

	body := w.Body.String()

	allocDesc := metricDescription("Alloc")
	expectedHelp := "# HELP Alloc " + allocDesc
	if !strings.Contains(body, expectedHelp) {
		t.Errorf("expected body to contain %q", expectedHelp)
	}
	if !strings.Contains(body, "# TYPE Alloc gauge") {
		t.Errorf("expected body to contain TYPE gauge for Alloc")
	}
	if !strings.Contains(body, "Alloc 1024.5") {
		t.Errorf("expected body to contain 'Alloc 1024.5'")
	}

	if !strings.Contains(body, "# TYPE PollCount counter") {
		t.Errorf("expected body to contain TYPE counter for PollCount")
	}
	if !strings.Contains(body, "PollCount 42") {
		t.Errorf("expected body to contain 'PollCount 42'")
	}
}

func updateJSON(t *testing.T, router *gin.Engine, m models.Metrics) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(m)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/update", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

func getJSON(t *testing.T, router *gin.Engine, m models.Metrics) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(m)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/value", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

func TestUpdateJSON_Gauge(t *testing.T) {
	router, store := setupRouter()

	v := 123.45
	w := updateJSON(t, router, models.Metrics{ID: "test_gauge", MType: "gauge", Value: &v})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := models.Metrics{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.ID != "test_gauge" || resp.MType != "gauge" || *resp.Value != 123.45 {
		t.Errorf("unexpected response: %+v", resp)
	}

	stored, ok, _ := store.GetGauge(context.Background(),"test_gauge")
	if !ok || stored != 123.45 {
		t.Errorf("expected stored 123.45, got %v, %v", stored, ok)
	}
}

func TestUpdateJSON_Counter(t *testing.T) {
	router, store := setupRouter()

	d := int64(5)
	w := updateJSON(t, router, models.Metrics{ID: "test_counter", MType: "counter", Delta: &d})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := models.Metrics{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.ID != "test_counter" || resp.MType != "counter" || *resp.Delta != 5 {
		t.Errorf("unexpected response: %+v", resp)
	}

	stored, ok, _ := store.GetCounter(context.Background(),"test_counter")
	if !ok || stored != 5 {
		t.Errorf("expected stored 5, got %v, %v", stored, ok)
	}
}

func TestUpdateJSON_InvalidJSON(t *testing.T) {
	router, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/update", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateJSON_InvalidName(t *testing.T) {
	router, _ := setupRouter()

	v := 1.0
	w := updateJSON(t, router, models.Metrics{ID: "", MType: "gauge", Value: &v})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateJSON_GaugeMissingValue(t *testing.T) {
	router, _ := setupRouter()

	w := updateJSON(t, router, models.Metrics{ID: "g", MType: "gauge"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateJSON_CounterMissingDelta(t *testing.T) {
	router, _ := setupRouter()

	w := updateJSON(t, router, models.Metrics{ID: "c", MType: "counter"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateJSON_InvalidType(t *testing.T) {
	router, _ := setupRouter()

	w := updateJSON(t, router, models.Metrics{ID: "m", MType: "unknown"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetJSON_Gauge(t *testing.T) {
	router, store := setupRouter()
	store.UpdateGauge(context.Background(),"gauge1", 42.5)

	w := getJSON(t, router, models.Metrics{ID: "gauge1", MType: "gauge"})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := models.Metrics{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.ID != "gauge1" || resp.MType != "gauge" {
		t.Errorf("unexpected id/type: %+v", resp)
	}
	if resp.Value == nil || *resp.Value != 42.5 {
		t.Errorf("expected value 42.5, got %+v", resp.Value)
	}
}

func TestGetJSON_Counter(t *testing.T) {
	router, store := setupRouter()
	store.UpdateCounter(context.Background(),"counter1", 99)

	w := getJSON(t, router, models.Metrics{ID: "counter1", MType: "counter"})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := models.Metrics{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.ID != "counter1" || resp.MType != "counter" {
		t.Errorf("unexpected id/type: %+v", resp)
	}
	if resp.Delta == nil || *resp.Delta != 99 {
		t.Errorf("expected delta 99, got %+v", resp.Delta)
	}
}

func TestGetJSON_NotFound(t *testing.T) {
	router, _ := setupRouter()

	w := getJSON(t, router, models.Metrics{ID: "nonexistent", MType: "gauge"})

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetJSON_InvalidJSON(t *testing.T) {
	router, _ := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/value", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetJSON_InvalidName(t *testing.T) {
	router, _ := setupRouter()

	w := getJSON(t, router, models.Metrics{ID: "", MType: "gauge"})

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
