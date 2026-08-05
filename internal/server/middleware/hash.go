package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/shukalov/go-ya/pkg/hash"
)

type hashWriter struct {
	gin.ResponseWriter
	key string
}

func (w *hashWriter) Write(data []byte) (int, error) {
	if w.key == "" {
		return w.ResponseWriter.Write(data)
	}
	w.Header().Set("HashSHA256", hash.Sign(data, w.key))
	return w.ResponseWriter.Write(data)
}

func Hash(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key == "" {
			c.Next()
			return
		}

		if hashHeader := c.GetHeader("HashSHA256"); hashHeader != "" {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))

			if hashHeader != hash.Sign(body, key) {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		c.Writer = &hashWriter{ResponseWriter: c.Writer, key: key}
		c.Next()
	}
}
