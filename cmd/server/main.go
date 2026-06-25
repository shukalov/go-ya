package main

import (
	"log"

	"github.com/shukalov/go-ya/internal/server"
	"github.com/shukalov/go-ya/internal/server/storage"
)

func main() {
	// Создаем хранилище
	storage := storage.NewMemStorage()

	// Создаем сервер
	srv := server.NewServer(":8080", storage)

	// Запускаем сервер
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
