package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/andypc/abn-news/internal/models"
	"github.com/andypc/abn-news/internal/store"
)

type SourceHandler struct {
	store *store.Store
}

func NewSourceHandler(s *store.Store) *SourceHandler {
	return &SourceHandler{store: s}
}

func (h *SourceHandler) Health(c *gin.Context) {
	health, err := h.store.GetSourceHealth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if health == nil {
		health = []models.SourceHealth(nil)
	}
	c.JSON(http.StatusOK, health)
}
