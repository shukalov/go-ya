package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shukalov/go-ya/internal/server/handlers"
	"github.com/shukalov/go-ya/internal/server/storage"
)

// Server - структура HTTP сервера
type Server struct {
	address string
	engine  *gin.Engine
}

// NewServer - создает новый сервер
func NewServer(address string, storage storage.Storage) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	h := handlers.NewMetricsHandler(storage)

	engine.POST("/update/:type/:name/:value", h.Update)
	engine.GET("/value/:type/:name", h.Get)
	engine.GET("/", h.Index)
	engine.GET("/metrics", h.Metrics)

	return &Server{
		address: address,
		engine:  engine,
	}
}

// Run - запускает сервер
func (s *Server) Run() error {
	fmt.Printf("Server is running on %s\n", s.address)
	return http.ListenAndServe(s.address, s.engine)
}
