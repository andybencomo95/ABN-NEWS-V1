package store

import (
	"context"
	"fmt"

	"github.com/andypc/abn-news/internal/models"
)

func (s *Store) ListCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, slug, name, description FROM categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.Description); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}
