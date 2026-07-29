package models

import "time"

type Article struct {
	ID          int64     `json:"id" db:"id"`
	SourceID    int64     `json:"source_id" db:"source_id"`
	CategoryID  *int64    `json:"category_id,omitempty" db:"category_id"`
	Title       string    `json:"title" db:"title"`
	Slug        string    `json:"slug" db:"slug"`
	URL         string    `json:"url" db:"url"`
	Content     string    `json:"content" db:"content"`
	Excerpt     string    `json:"excerpt,omitempty" db:"excerpt"`
	ImageURL    string    `json:"image_url,omitempty" db:"image_url"`
	Author      string    `json:"author,omitempty" db:"author"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	FetchedAt   time.Time `json:"fetched_at" db:"fetched_at"`
	Hash        string    `json:"-" db:"hash"`
}

type PaginatedArticles struct {
	Articles   []Article `json:"articles"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	Total      int       `json:"total"`
	TotalPages int       `json:"total_pages"`
}

type CreateArticleInput struct {
	SourceID    int64
	CategoryID  *int64
	Title       string
	Slug        string
	URL         string
	Content     string
	Excerpt     string
	ImageURL    string
	Author      string
	PublishedAt time.Time
}
