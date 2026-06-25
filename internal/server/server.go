package server

import (
	"fmt"
	"net/http"

	"github.com/shukalov/go-ya/internal/server/handlers"
	"github.com/shukalov/go-ya/internal/server/storage"
)

// Server - структура HTTP сервера
type Server struct {
	address string
	storage storage.Storage
	handler *handlers.MetricsHandler
}

// NewServer - создает новый сервер
func NewServer(address string, storage storage.Storage) *Server {
	return &Server{
		address: address,
		storage: storage,
		handler: handlers.NewMetricsHandler(storage),
	}
}

// RegisterHandlers - регистрирует обработчики
func (s *Server) RegisterHandlers() {
	http.HandleFunc("/update/", s.handler.Update)
	http.HandleFunc("/value/", s.handler.Get)
}

// Run - запускает сервер
func (s *Server) Run() error {
	s.RegisterHandlers()
	fmt.Printf("Server is running on %s\n", s.address)
	return http.ListenAndServe(s.address, nil)
}
