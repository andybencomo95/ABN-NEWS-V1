package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/andypc/abn-news/internal/models"
	"github.com/andypc/abn-news/internal/rss"
	"github.com/andypc/abn-news/internal/store"
)

type ImageRefetchHandler struct {
	store   *store.Store
	wiki    *rss.WikipediaClient
}

func NewImageRefetchHandler(s *store.Store) *ImageRefetchHandler {
	return &ImageRefetchHandler{
		store: s,
		wiki:  rss.NewWikipediaClient(),
	}
}

// RefetchImages re-fetches images for articles that have the default category image.
// This is a background job that runs asynchronously.
func (h *ImageRefetchHandler) RefetchImages(c *gin.Context) {
	// Get articles with default images (limited to avoid overwhelming APIs)
	articles, err := h.store.ListArticlesWithDefaultImages(c.Request.Context(), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(articles) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "no articles need image refetch", "count": 0})
		return
	}

	// Process asynchronously
	go h.processRefetch(articles)

	c.JSON(http.StatusOK, gin.H{
		"message": "image refetch started",
		"count":   len(articles),
	})
}

func (h *ImageRefetchHandler) processRefetch(articles []models.Article) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3) // max 3 concurrent searches

	for _, article := range articles {
		wg.Add(1)
		sem <- struct{}{}
		go func(a models.Article) {
			defer wg.Done()
			defer func() { <-sem }()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			imgURL, err := h.wiki.FetchImage(ctx, a.Title)
			if err != nil || imgURL == "" {
				return
			}

			_ = h.store.UpdateArticleImage(ctx, a.ID, imgURL)
		}(article)
	}
	wg.Wait()
}
