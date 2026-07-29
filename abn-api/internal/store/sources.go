package store

import (
	"context"
	"fmt"
	"time"

	"github.com/andypc/abn-news/internal/models"
)

func (s *Store) ListSources(ctx context.Context) ([]models.Source, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, site_url, feed_url, category_id, interval_sec, health_status, last_fetched_at, last_error, backoff_until FROM sources ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var src models.Source
		if err := rows.Scan(&src.ID, &src.Name, &src.SiteURL, &src.FeedURL, &src.CategoryID, &src.IntervalSec, &src.HealthStatus, &src.LastFetched, &src.LastError, &src.BackoffUntil); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, src)
	}
	return sources, nil
}

func (s *Store) GetActiveSources(ctx context.Context) ([]models.Source, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, site_url, feed_url, category_id, interval_sec, health_status, last_fetched_at, last_error, backoff_until
		FROM sources
		WHERE (backoff_until IS NULL OR backoff_until <= NOW())
		ORDER BY last_fetched_at ASC NULLS FIRST
	`)
	if err != nil {
		return nil, fmt.Errorf("get active sources: %w", err)
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var src models.Source
		if err := rows.Scan(&src.ID, &src.Name, &src.SiteURL, &src.FeedURL, &src.CategoryID, &src.IntervalSec, &src.HealthStatus, &src.LastFetched, &src.LastError, &src.BackoffUntil); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, src)
	}
	return sources, nil
}

func (s *Store) UpdateSourceHealth(ctx context.Context, sourceID int64, status string, lastError string, backoffUntil *time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sources SET health_status = $1, last_error = $2, backoff_until = $3, last_fetched_at = NOW()
		WHERE id = $4
	`, status, lastError, backoffUntil, sourceID)
	if err != nil {
		return fmt.Errorf("update source health: %w", err)
	}
	return nil
}

func (s *Store) RecordFetchSuccess(ctx context.Context, sourceID int64, articlesCount int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sources SET health_status = 'ok', last_error = NULL, backoff_until = NULL, last_fetched_at = NOW()
		WHERE id = $1
	`, sourceID)
	if err != nil {
		return fmt.Errorf("record fetch success: %w", err)
	}
	return nil
}

func (s *Store) GetSourceHealth(ctx context.Context) ([]models.SourceHealth, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.name, s.health_status, s.last_fetched_at, s.backoff_until, COALESCE(s.last_error, ''),
			   (SELECT COUNT(*) FROM articles WHERE source_id = s.id AND fetched_at > NOW() - INTERVAL '1 day') as articles_today
		FROM sources s ORDER BY s.name
	`)
	if err != nil {
		return nil, fmt.Errorf("get source health: %w", err)
	}
	defer rows.Close()

	var health []models.SourceHealth
	for rows.Next() {
		var h models.SourceHealth
		if err := rows.Scan(&h.Name, &h.Status, &h.LastFetched, &h.BackoffUntil, &h.LastError, &h.ArticlesToday); err != nil {
			return nil, fmt.Errorf("scan source health: %w", err)
		}
		health = append(health, h)
	}
	return health, nil
}
