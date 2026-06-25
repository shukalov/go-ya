package sender

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shukalov/go-ya/pkg/models"
)

func TestNewHTTPSender(t *testing.T) {
	sender := NewHTTPSender("http://localhost:8080")
	if sender == nil {
		t.Error("expected sender to be created")
	}
	if sender.serverAddress != "http://localhost:8080" {
		t.Errorf("expected serverAddress http://localhost:8080, got %s", sender.serverAddress)
	}
	if sender.client == nil {
		t.Error("expected client to be initialized")
	}
	if sender.client.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", sender.client.Timeout)
	}
}

func TestHTTPSender_SendMetric_Success(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// Проверяем Content-Type
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("expected text/plain, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL)
	err := sender.SendMetric(models.TypeGauge, "test", 123.45)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHTTPSender_SendMetric_Error(t *testing.T) {
	// Создаем тестовый сервер, который возвращает ошибку
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL)
	err := sender.SendMetric(models.TypeGauge, "test", 123.45)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestHTTPSender_SendMetric_InvalidValue(t *testing.T) {
	sender := NewHTTPSender("http://localhost:8080")

	// Передаем неподдерживаемый тип
	err := sender.SendMetric(models.TypeGauge, "test", "invalid")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestHTTPSender_SendAllMetrics(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL)

	metrics := models.RuntimeMetrics{
		Alloc:       1024.5,
		HeapAlloc:   512.3,
		PollCount:   5,
		RandomValue: 42.0,
	}

	err := sender.SendAllMetrics(metrics, 5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHTTPSender_SendAllMetrics_Error(t *testing.T) {
	// Создаем тестовый сервер с ошибкой
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL)

	metrics := models.RuntimeMetrics{
		Alloc: 1024.5,
	}

	err := sender.SendAllMetrics(metrics, 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestHTTPSender_Timeout(t *testing.T) {
	// Создаем медленный сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL)
	sender.client.Timeout = 1 * time.Second

	err := sender.SendMetric(models.TypeGauge, "test", 123.45)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}
