package server

import (
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/server/handlers"
	"github.com/shukalov/go-ya/internal/server/middleware"
	"github.com/shukalov/go-ya/internal/server/storage"
)

type Server struct {
	address string
	engine  *gin.Engine
	logger  *zap.Logger
}

func NewServer(address string, storage storage.Storage, logger *zap.Logger) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.Logging(logger))
	engine.Use(gzip.DefaultDecompressHandle)
	engine.Use(middleware.Compress())

	h := handlers.NewMetricsHandler(storage)

	engine.POST("/update/:type/:name/:value", h.Update)
	engine.POST("/update", h.UpdateJSON)
	engine.GET("/value/:type/:name", h.Get)
	engine.POST("/value", h.GetJSON)
	engine.GET("/", h.Index)
	engine.GET("/metrics", h.Metrics)

	return &Server{
		address: address,
		engine:  engine,
		logger:  logger,
	}
}

func (s *Server) Run() error {
	s.logger.Info("server starting", zap.String("address", s.address))
	return http.ListenAndServe(s.address, s.engine)
}
