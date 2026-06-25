package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	config := Config{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerAddress:  "http://localhost:8080",
	}

	agent := NewAgent(config)
	if agent == nil {
		t.Error("expected agent to be created")
	}
	if agent.config.PollInterval != 2*time.Second {
		t.Errorf("expected PollInterval 2s, got %v", agent.config.PollInterval)
	}
	if agent.collector == nil {
		t.Error("expected collector to be initialized")
	}
	if agent.sender == nil {
		t.Error("expected sender to be initialized")
	}
}

func TestAgent_CollectMetrics(t *testing.T) {
	config := Config{
		PollInterval:   1 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerAddress:  "http://localhost:8080",
	}

	agent := NewAgent(config)

	// Собираем метрики вручную
	agent.collector.Collect()

	metrics := agent.collector.GetMetrics()
	if metrics.Counters["PollCount"] != 1 {
		t.Errorf("expected PollCount 1, got %d", metrics.Counters["PollCount"])
	}
}

func TestAgent_SendMetrics(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := Config{
		PollInterval:   1 * time.Second,
		ReportInterval: 2 * time.Second,
		ServerAddress:  server.URL,
	}

	agent := NewAgent(config)

	// Собираем метрики
	agent.collector.Collect()

	// Отправляем метрики
	metrics := agent.collector.GetMetrics()

	err := agent.sender.SendAllMetrics(metrics)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgent_Run(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := Config{
		PollInterval:   100 * time.Millisecond,
		ReportInterval: 200 * time.Millisecond,
		ServerAddress:  server.URL,
	}

	agent := NewAgent(config)

	// Запускаем агента в горутине
	done := make(chan bool)
	go func() {
		agent.Run()
		done <- true
	}()

	// Ждем несколько циклов
	time.Sleep(500 * time.Millisecond)

	// Проверяем, что метрики собираются
	metrics := agent.collector.GetMetrics()
	if metrics.Counters["PollCount"] < 3 {
		t.Errorf("expected PollCount >= 3, got %d", metrics.Counters["PollCount"])
	}
}
