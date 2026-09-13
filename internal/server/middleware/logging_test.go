package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zaptest"
)

func TestLogging(t *testing.T) {
	logger := zaptest.NewLogger(t)

	rr := httptest.NewRecorder()
	r := gin.New()
	r.Use(Logging(logger))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestLogging_PropagatesResponse(t *testing.T) {
	logger := zaptest.NewLogger(t)

	rr := httptest.NewRecorder()
	r := gin.New()
	r.Use(Logging(logger))
	r.GET("/hello", func(c *gin.Context) { c.String(200, "hello") })

	req, _ := http.NewRequest(http.MethodGet, "/hello", nil)
	r.ServeHTTP(rr, req)

	if rr.Body.String() != "hello" {
		t.Errorf("expected 'hello', got %s", rr.Body.String())
	}
}
