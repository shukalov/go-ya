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
	logger := zap.NewNop()
	store := storage.NewMemStorage()
	return NewServer("localhost:8080", store, logger, dsn)
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

	if err := srv.runDB(); err != nil {
		t.Fatalf("runDB failed: %v", err)
	}
	defer srv.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	srv.engine.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
