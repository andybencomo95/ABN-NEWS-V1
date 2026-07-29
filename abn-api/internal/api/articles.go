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
		translateArticleList(c.Request.Context(), h.translator, articles.Articles)
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
		translateSingleArticle(c.Request.Context(), h.translator, article)
	}

	c.JSON(http.StatusOK, article)
}

func translateArticleList(ctx context.Context, t *translate.Client, articles []models.Article) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // max 5 concurrent translations

	for i := range articles {
		wg.Add(1)
		sem <- struct{}{}
		go func(a *models.Article) {
			defer wg.Done()
			defer func() { <-sem }()
			translateSingleArticle(ctx, t, a)
		}(&articles[i])
	}
	wg.Wait()
}

func translateSingleArticle(ctx context.Context, t *translate.Client, a *models.Article) {
	if a.Title != "" {
		if title, err := t.Translate(ctx, a.Title, "en", "es"); err == nil {
			a.Title = title
		}
	}
	if a.Excerpt != "" {
		if excerpt, err := t.Translate(ctx, a.Excerpt, "en", "es"); err == nil {
			a.Excerpt = excerpt
		}
	}
	// Skip full content translation for list views (only translate title+excerpt)
}
