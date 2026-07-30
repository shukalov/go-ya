package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shukalov/go-ya/internal/server/storage"
	"go.uber.org/zap"
)

func setupTestServer(dsn string) *Server {
	gin.SetMode(gin.TestMode)
	srv := NewServer(Config{
		Address:     "localhost:8080",
		Logger:      zap.NewNop(),
		DatabaseDSN: dsn,
	})
	srv.store = storage.NewMemStorage()
	srv.routes()
	return srv
}

func TestPingHandler_NoDSN(t *testing.T) {
	srv := setupTestServer("")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	srv.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestPingHandler_InvalidDSN(t *testing.T) {
	srv := setupTestServer("postgres://invalid:invalid@localhost:9999/invalid")

	if err := srv.initDB(); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}
	defer srv.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	srv.engine.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
