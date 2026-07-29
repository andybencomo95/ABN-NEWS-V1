package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/andypc/abn-news/internal/models"
	"github.com/andypc/abn-news/internal/store"
)

type CategoryHandler struct {
	store *store.Store
}

func NewCategoryHandler(s *store.Store) *CategoryHandler {
	return &CategoryHandler{store: s}
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.store.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if categories == nil {
		categories = []models.Category(nil)
	}
	c.JSON(http.StatusOK, categories)
}
