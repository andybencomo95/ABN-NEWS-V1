package store

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/andypc/abn-news/internal/models"
)

const articlesPerPage = 10

func (s *Store) CreateArticle(ctx context.Context, input models.CreateArticleInput) (*models.Article, error) {
	slug := createSlug(input.Title)
	hash := createHash(input.URL, input.PublishedAt)

	var a models.Article
	err := s.pool.QueryRow(ctx, `
		INSERT INTO articles (source_id, category_id, title, slug, url, content, excerpt, image_url, author, published_at, hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (hash) DO NOTHING
		RETURNING id, source_id, category_id, title, slug, url, content, excerpt, image_url, author, published_at, fetched_at, hash
	`, input.SourceID, input.CategoryID, input.Title, slug, input.URL, input.Content, input.Excerpt, input.ImageURL, input.Author, input.PublishedAt, hash).Scan(
		&a.ID, &a.SourceID, &a.CategoryID, &a.Title, &a.Slug, &a.URL,
		&a.Content, &a.Excerpt, &a.ImageURL, &a.Author, &a.PublishedAt, &a.FetchedAt, &a.Hash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrDuplicateHash
		}
		return nil, fmt.Errorf("create article: %w", err)
	}
	return &a, nil
}

func (s *Store) ListArticles(ctx context.Context, categorySlug string, page, limit int) (*models.PaginatedArticles, error) {
	if limit <= 0 || limit > 50 {
		limit = articlesPerPage
	}
	offset := (page - 1) * limit

	var total int
	var articles []models.Article
	var err error

	// 'general' shows all articles (catch-all category)
	if categorySlug != "" && categorySlug != "general" {
		err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM articles a JOIN categories c ON a.category_id = c.id WHERE c.slug = $1`, categorySlug).Scan(&total)
		if err != nil {
			return nil, fmt.Errorf("count articles: %w", err)
		}
		rows, err := s.pool.Query(ctx, `
			SELECT a.id, a.source_id, a.category_id, a.title, a.slug, a.url,
				   COALESCE(a.content, '') as content, COALESCE(a.excerpt, '') as excerpt,
				   COALESCE(a.image_url, '') as image_url, COALESCE(a.author, '') as author,
				   a.published_at, a.fetched_at
			FROM articles a JOIN categories c ON a.category_id = c.id
			WHERE c.slug = $1
			ORDER BY a.published_at DESC LIMIT $2 OFFSET $3
		`, categorySlug, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("query articles: %w", err)
		}
		articles, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.Article])
		if err != nil {
			return nil, fmt.Errorf("collect articles: %w", err)
		}
	} else {
		err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM articles`).Scan(&total)
		if err != nil {
			return nil, fmt.Errorf("count articles: %w", err)
		}
		rows, err := s.pool.Query(ctx, `
			SELECT id, source_id, category_id, title, slug, url,
				   COALESCE(content, '') as content, COALESCE(excerpt, '') as excerpt,
				   COALESCE(image_url, '') as image_url, COALESCE(author, '') as author,
				   published_at, fetched_at
			FROM articles ORDER BY published_at DESC LIMIT $1 OFFSET $2
		`, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("query articles: %w", err)
		}
		articles, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.Article])
		if err != nil {
			return nil, fmt.Errorf("collect articles: %w", err)
		}
	}

	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	return &models.PaginatedArticles{
		Articles:   articles,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *Store) GetArticleBySlug(ctx context.Context, slug string) (*models.Article, error) {
	var a models.Article
	err := s.pool.QueryRow(ctx, `
		SELECT id, source_id, category_id, title, slug, url,
			   COALESCE(content, '') as content, COALESCE(excerpt, '') as excerpt,
			   COALESCE(image_url, '') as image_url, COALESCE(author, '') as author,
			   published_at, fetched_at
		FROM articles WHERE slug = $1
	`, slug).Scan(
		&a.ID, &a.SourceID, &a.CategoryID, &a.Title, &a.Slug, &a.URL,
		&a.Content, &a.Excerpt, &a.ImageURL, &a.Author, &a.PublishedAt, &a.FetchedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get article by slug: %w", err)
	}
	return &a, nil
}

func (s *Store) UpdateArticleImage(ctx context.Context, id int64, imageURL string) error {
	_, err := s.pool.Exec(ctx, `UPDATE articles SET image_url = $1 WHERE id = $2`, imageURL, id)
	if err != nil {
		return fmt.Errorf("update article image: %w", err)
	}
	return nil
}

// ListArticlesWithDefaultImages returns articles that have the default category image,
// so they can be re-fetched with better images.
var defaultImagePatterns = []string{
	"%Dampfturbine%",
	"%Land_on_the_Moon%",
	"%Youth-soccer%",
	"%Map_of_countries%",
	"%Bisonte%",
	"%DNA_simple%",
}

func (s *Store) ListArticlesWithDefaultImages(ctx context.Context, limit int) ([]models.Article, error) {
	if limit <= 0 {
		limit = 50
	}

	// Build WHERE clause for default images
	conditions := make([]string, len(defaultImagePatterns))
	for i, pattern := range defaultImagePatterns {
		conditions[i] = "a.image_url LIKE '" + pattern + "'"
	}
	whereClause := ""
	for i, cond := range conditions {
		if i == 0 {
			whereClause = cond
		} else {
			whereClause += " OR " + cond
		}
	}

	query := `
		SELECT a.id, a.source_id, a.category_id, a.title, a.slug, a.url,
			   COALESCE(a.content, '') as content, COALESCE(a.excerpt, '') as excerpt,
			   COALESCE(a.image_url, '') as image_url, COALESCE(a.author, '') as author,
			   a.published_at, a.fetched_at
		FROM articles a
		WHERE ` + whereClause + `
		ORDER BY a.published_at DESC
		LIMIT $1
	`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list articles with default images: %w", err)
	}
	defer rows.Close()

	articles, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.Article])
	if err != nil {
		return nil, fmt.Errorf("collect articles: %w", err)
	}
	return articles, nil
}

func createSlug(title string) string {
	lower := strings.ToLower(title)
	// Replace non-alphanumeric with hyphens
	var b strings.Builder
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else if r == ' ' || r == '_' {
			b.WriteRune('-')
		}
	}
	slug := b.String()
	// Collapse multiple hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if len(slug) == 0 {
		slug = fmt.Sprintf("%x", sha256.Sum256([]byte(title)))[:12]
	}
	if len(slug) > 100 {
		slug = slug[:100]
	}
	return slug
}

func createHash(url string, publishedAt interface{}) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%v", url, publishedAt)))
	return fmt.Sprintf("%x", h)
}
