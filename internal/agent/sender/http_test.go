package sender

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shukalov/go-ya/pkg/models"
)

func TestNewHTTPSender(t *testing.T) {
	sender := NewHTTPSender("http://localhost:8080", "")
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Errorf("expected gzip, got %s", r.Header.Get("Content-Encoding"))
		}
		if r.URL.Path != "/update" {
			t.Errorf("expected /update, got %s", r.URL.Path)
		}

		var reader io.Reader = r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				t.Errorf("failed to create gzip reader: %v", err)
			}
			defer gz.Close()
			reader = gz
		}

		body, _ := io.ReadAll(reader)
		var m models.Metrics
		if err := json.Unmarshal(body, &m); err != nil {
			t.Errorf("failed to unmarshal body: %v", err)
		}
		if m.ID != "test" {
			t.Errorf("expected id test, got %s", m.ID)
		}
		if m.MType != "gauge" {
			t.Errorf("expected type gauge, got %s", m.MType)
		}
		if m.Value == nil || *m.Value != 123.45 {
			t.Errorf("expected value 123.45, got %v", m.Value)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendMetric("gauge", "test", 123.45)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHTTPSender_SendMetric_Counter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader = r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				t.Errorf("failed to create gzip reader: %v", err)
			}
			defer gz.Close()
			reader = gz
		}

		body, _ := io.ReadAll(reader)
		var m models.Metrics
		if err := json.Unmarshal(body, &m); err != nil {
			t.Errorf("failed to unmarshal body: %v", err)
		}
		if m.Delta == nil || *m.Delta != 5 {
			t.Errorf("expected delta 5, got %v", m.Delta)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendMetric("counter", "PollCount", int64(5))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHTTPSender_SendMetric_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendMetric("gauge", "test", 123.45)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestHTTPSender_SendMetric_InvalidValue(t *testing.T) {
	sender := NewHTTPSender("http://localhost:8080", "")

	err := sender.SendMetric("gauge", "test", "invalid")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestHTTPSender_SendAllMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")

	metrics := models.RuntimeMetrics{
		Gauges:   map[string]float64{"Alloc": 1024.5, "HeapAlloc": 512.3, "RandomValue": 42.0},
		Counters: map[string]int64{"PollCount": 5},
	}

	err := sender.SendAllMetrics(metrics)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHTTPSender_SendAllMetrics_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")

	metrics := models.RuntimeMetrics{
		Gauges: map[string]float64{"Alloc": 1024.5},
	}

	err := sender.SendAllMetrics(metrics)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestHTTPSender_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	sender.client.Timeout = 1 * time.Second

	err := sender.SendMetric("gauge", "test", 123.45)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}
