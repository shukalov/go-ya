package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type compressWriter struct {
	gin.ResponseWriter
	gz         *gzip.Writer
	compressed bool
}

func (w *compressWriter) Write(data []byte) (int, error) {
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") && !strings.HasPrefix(ct, "text/html") {
		return w.ResponseWriter.Write(data)
	}

	status := w.ResponseWriter.Status()
	if status >= http.StatusMultipleChoices {
		return w.ResponseWriter.Write(data)
	}

	w.Header().Del("Content-Length")
	w.Header().Set("Content-Encoding", "gzip")

	if w.gz == nil {
		w.gz = gzip.NewWriter(w.ResponseWriter)
	}

	w.compressed = true
	return w.gz.Write(data)
}

func Compress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		cw := &compressWriter{ResponseWriter: c.Writer}
		c.Writer = cw

		defer func() {
			if cw.compressed {
				_ = cw.gz.Close()
			}
		}()

		c.Next()
	}
}
