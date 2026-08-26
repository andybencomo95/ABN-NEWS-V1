package api

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/andypc/abn-news/internal/models"
	"github.com/andypc/abn-news/internal/store"
	"github.com/andypc/abn-news/internal/translate"
)

type ArticleHandler struct {
	store      *store.Store
	translator *translate.Client
}

func NewArticleHandler(s *store.Store, t *translate.Client) *ArticleHandler {
	return &ArticleHandler{store: s, translator: t}
}

func (h *ArticleHandler) List(c *gin.Context) {
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	lang := c.Query("lang")

	if page < 1 {
		page = 1
	}

	articles, err := h.store.ListArticles(c.Request.Context(), category, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if lang == "es" && h.translator != nil {
		if err := translateArticleList(c.Request.Context(), h.translator, articles.Articles); err != nil {
			c.Header("X-Cache-Skip", "1")
		}
	}

	c.JSON(http.StatusOK, articles)
}

func (h *ArticleHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	lang := c.DefaultQuery("lang", "")

	article, err := h.store.GetArticleBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	if lang == "es" && h.translator != nil {
		if err := translateSingleArticle(c.Request.Context(), h.translator, article); err != nil {
			c.Header("X-Cache-Skip", "1")
		}
	}

	c.JSON(http.StatusOK, article)
}

func translateArticleList(ctx context.Context, t *translate.Client, articles []models.Article) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // max 5 concurrent translations

	var mu sync.Mutex
	var firstErr error

	for i := range articles {
		wg.Add(1)
		sem <- struct{}{}
		go func(a *models.Article) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := translateSingleArticle(ctx, t, a); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(&articles[i])
	}
	wg.Wait()
	return firstErr
}

func translateSingleArticle(ctx context.Context, t *translate.Client, a *models.Article) error {
	var firstErr error
	if a.Title != "" {
		title, err := t.Translate(ctx, a.Title, "en", "es")
		if err != nil {
			firstErr = err
		} else {
			a.Title = title
		}
	}
	if a.Excerpt != "" {
		excerpt, err := t.Translate(ctx, a.Excerpt, "en", "es")
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		} else {
			a.Excerpt = excerpt
		}
	}
	// Skip full content translation for list views (only translate title+excerpt)
	return firstErr
}
