package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/server/handlers"
	"github.com/shukalov/go-ya/internal/server/middleware"
	"github.com/shukalov/go-ya/internal/server/storage"
)

type Server struct {
	address string
	engine  *gin.Engine
	logger  *zap.Logger
	db      *sql.DB
	dsn     string
}

func NewServer(address string, storage storage.Storage, logger *zap.Logger, dsn string) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		address: address,
		logger:  logger,
		dsn:     dsn,
	}

	s.engine = gin.New()
	s.engine.Use(gin.Recovery())
	s.engine.Use(middleware.Logging(logger))
	s.engine.Use(gzip.DefaultDecompressHandle)
	s.engine.Use(middleware.Compress())

	h := handlers.NewMetricsHandler(storage)

	s.engine.POST("/update/:type/:name/:value", h.Update)
	s.engine.POST("/update", h.UpdateJSON)
	s.engine.GET("/value/:type/:name", h.Get)
	s.engine.POST("/value", h.GetJSON)
	s.engine.GET("/", h.Index)
	s.engine.GET("/metrics", h.Metrics)

	s.engine.GET("/ping", func(c *gin.Context) {
		if s.db == nil {
			c.String(http.StatusOK, "OK")
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()
		if err := s.db.PingContext(ctx); err != nil {
			c.String(http.StatusInternalServerError, "database connection failed")
			return
		}
		c.String(http.StatusOK, "OK")
	})

	return s
}

func (s *Server) runDB() error {
	db, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	s.db = db
	return nil
}

func (s *Server) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Server) Run() error {
	s.logger.Info("server starting", zap.String("address", s.address))
	if s.dsn != "" {
		if err := s.runDB(); err != nil {
			return err
		}
		defer s.Close()
	}
	return http.ListenAndServe(s.address, s.engine)
}
