package main

import (
	"log"
	"time"

	"github.com/shukalov/go-ya/internal/agent"
)

func main() {
	// Конфигурация агента
	config := agent.Config{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerAddress:  "http://localhost:8080",
	}

	// Создаем агента
	agent := agent.NewAgent(config)

	// Запускаем агента
	if err := agent.Run(); err != nil {
		log.Fatalf("Agent error: %v", err)
	}
}
