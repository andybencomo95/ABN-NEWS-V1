package models

import "time"

type Source struct {
	ID           int64      `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	SiteURL      string     `json:"site_url" db:"site_url"`
	FeedURL      string     `json:"feed_url" db:"feed_url"`
	CategoryID   *int64     `json:"category_id,omitempty" db:"category_id"`
	IntervalSec  int        `json:"interval_sec" db:"interval_sec"`
	HealthStatus string     `json:"health_status" db:"health_status"`
	LastFetched  *time.Time `json:"last_fetched_at,omitempty" db:"last_fetched_at"`
	LastError    *string    `json:"last_error,omitempty" db:"last_error"`
	BackoffUntil *time.Time `json:"backoff_until,omitempty" db:"backoff_until"`
}

type SourceHealth struct {
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	LastFetched   *time.Time `json:"last_fetched_at,omitempty"`
	BackoffUntil  *time.Time `json:"backoff_until,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
	ArticlesToday int        `json:"articles_today"`
}
