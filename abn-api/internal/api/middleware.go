package api

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/andypc/abn-news/internal/cache"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

type cacheWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *cacheWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func CacheMiddleware(cache *cache.Cache, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		key := "api:" + c.Request.URL.RequestURI()
		ctx := c.Request.Context()

		// Try Redis cache
		var cached []byte
		if err := cache.Get(ctx, key, &cached); err == nil {
			c.Data(http.StatusOK, "application/json; charset=utf-8", cached)
			c.Abort()
			return
		}

		// Capture response
		w := &cacheWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = w

		c.Next()

		// Store in cache only if successful, not empty, and the handler didn't
		// request to skip caching (e.g. a response with failed translations).
		if c.Writer.Status() == http.StatusOK && len(w.body.Bytes()) > 10 && c.Writer.Header().Get("X-Cache-Skip") == "" {
			_ = cache.Set(ctx, key, w.body.Bytes(), ttl)
		}
	}
}
