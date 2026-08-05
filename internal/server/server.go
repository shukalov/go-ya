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
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/shukalov/go-ya/internal/server/handlers"
	"github.com/shukalov/go-ya/internal/server/middleware"
	"github.com/shukalov/go-ya/internal/server/storage"
	"github.com/shukalov/go-ya/internal/server/storage/migrations"
)

type Config struct {
	Address       string
	Logger        *zap.Logger
	DatabaseDSN   string
	FilePath      string
	StoreInterval time.Duration
	Restore       bool
	Key           string
}

type Server struct {
	cfg    Config
	engine *gin.Engine
	db     *sql.DB
	store  storage.Storage
}

func NewServer(cfg Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{cfg: cfg}

	s.engine = gin.New()
	s.engine.Use(gin.Recovery())
	s.engine.Use(middleware.Logging(cfg.Logger))
	s.engine.Use(gzip.DefaultDecompressHandle)
	s.engine.Use(middleware.Compress())
	s.engine.Use(middleware.Hash(cfg.Key))

	return s
}

func (s *Server) initDB() error {
	db, err := sql.Open("pgx", s.cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	s.db = db
	return nil
}

func (s *Server) Migrate() error {
	goose.SetBaseFS(migrations.EmbedFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(s.db, ".")
}

func (s *Server) routes() {
	h := handlers.NewMetricsHandler(s.store)

	s.engine.POST("/update/:type/:name/:value/", h.Update)
	s.engine.POST("/update/", h.UpdateJSON)
	s.engine.POST("/updates/", h.UpdateJSONBatch)
	s.engine.GET("/value/:type/:name/", h.Get)
	s.engine.POST("/value/", h.GetJSON)
	s.engine.GET("/", h.Index)
	s.engine.GET("/metrics/", h.Metrics)

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
}

func (s *Server) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Server) Run() error {
	s.cfg.Logger.Info("server starting", zap.String("address", s.cfg.Address))
	defer s.Close()

	switch {
	case s.cfg.DatabaseDSN != "":
		if err := s.initDB(); err != nil {
			return err
		}
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
		s.store = storage.NewDBStorage(s.db, s.cfg.StoreInterval)
		s.cfg.Logger.Info("using database storage")

	case s.cfg.FilePath != "":
		s.store = storage.NewFileStorage(s.cfg.FilePath, s.cfg.StoreInterval)
		s.cfg.Logger.Info("using file storage")

	default:
		s.store = storage.NewMemStorage()
		s.cfg.Logger.Info("using memory storage")
	}

	if s.cfg.Restore {
		if err := s.store.Load(); err != nil {
			s.cfg.Logger.Error("failed to restore metrics", zap.Error(err))
		}
	}

	go s.store.Run()

	s.routes()
	return http.ListenAndServe(s.cfg.Address, s.engine)
}
