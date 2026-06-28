package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCompress_SkipWithoutAcceptEncoding(t *testing.T) {
	rr := httptest.NewRecorder()
	r := gin.New()
	r.Use(Compress())
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(200, `{"key":"value"}`)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected no gzip without Accept-Encoding")
	}
	if rr.Body.String() != `{"key":"value"}` {
		t.Errorf("expected uncompressed body, got %s", rr.Body.String())
	}
}

func TestCompress_JSON(t *testing.T) {
	rr := httptest.NewRecorder()
	r := gin.New()
	r.Use(Compress())
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(200, `{"key":"value"}`)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected gzip encoding")
	}

	body := decompress(t, rr.Body)
	if string(body) != `{"key":"value"}` {
		t.Errorf("expected {'key':'value'}, got %s", body)
	}
}

func TestCompress_HTML(t *testing.T) {
	rr := httptest.NewRecorder()
	r := gin.New()
	r.Use(Compress())
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, "<html></html>")
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected gzip encoding")
	}

	body := decompress(t, rr.Body)
	if string(body) != "<html></html>" {
		t.Errorf("expected '<html></html>', got %s", body)
	}
}

func TestCompress_PlainText(t *testing.T) {
	rr := httptest.NewRecorder()
	r := gin.New()
	r.Use(Compress())
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain")
		c.String(200, "ok")
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected no gzip for text/plain")
	}
	if rr.Body.String() != "ok" {
		t.Errorf("expected 'ok', got %s", rr.Body.String())
	}
}

func TestCompress_ErrorStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	r := gin.New()
	r.Use(Compress())
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(400, `{"error":"bad"}`)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected no gzip for 400 status")
	}
}

func TestCompress_GzipFooterIsValid(t *testing.T) {
	rr := httptest.NewRecorder()
	r := gin.New()
	r.Use(Compress())
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(200, `{"key":"value"}`)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected gzip encoding")
	}

	gzReader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}

	_, err = io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}

	if err := gzReader.Close(); err != nil {
		t.Fatalf("gzip footer invalid: %v", err)
	}
}

func decompress(t testing.TB, b *bytes.Buffer) []byte {
	t.Helper()
	gzReader, err := gzip.NewReader(b)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer gzReader.Close()

	body, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	return body
}
