package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/andypc/abn-news/internal/cache"
	"github.com/andypc/abn-news/internal/store"
	"github.com/andypc/abn-news/internal/translate"
)

func NewRouter(st *store.Store, cache *cache.Cache, translator *translate.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	articles := NewArticleHandler(st, translator)
	categories := NewCategoryHandler(st)
	sources := NewSourceHandler(st)

	apiGroup := r.Group("/api")
	apiGroup.Use(CacheMiddleware(cache, 30*time.Minute))
	{
		apiGroup.GET("/articles", articles.List)
		apiGroup.GET("/articles/:slug", articles.GetBySlug)
		apiGroup.GET("/categories", categories.List)
	}

	// Health endpoint (no cache)
	r.GET("/api/sources/health", sources.Health)

	return r
}
