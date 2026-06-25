package main

import (
	"flag"
	"log"

	"github.com/shukalov/go-ya/internal/server"
	"github.com/shukalov/go-ya/internal/server/storage"
)

func main() {
	addr := flag.String("a", "localhost:8080", "address endpoint")

	flag.Parse()

	storage := storage.NewMemStorage()
	srv := server.NewServer(*addr, storage)

	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
