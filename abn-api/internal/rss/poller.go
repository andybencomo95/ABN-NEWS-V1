package rss

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/andypc/abn-news/internal/models"
	"github.com/andypc/abn-news/internal/store"
)

type Poller struct {
	store   *store.Store
	fetcher *Fetcher
	logger  *slog.Logger
	wg      sync.WaitGroup
	cancel  context.CancelFunc
	wiki    *WikipediaClient
	breakers sync.Map // sourceID -> *CircuitBreaker
}

func NewPoller(s *store.Store, f *Fetcher, logger *slog.Logger) *Poller {
	return &Poller{
		store:   s,
		fetcher: f,
		logger:  logger,
		wiki:    NewWikipediaClient(),
	}
}

func (p *Poller) Start(ctx context.Context) {
	ctx, p.cancel = context.WithCancel(ctx)
	p.wg.Add(1)
	go p.loop(ctx)
}

func (p *Poller) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
}

func (p *Poller) loop(ctx context.Context) {
	defer p.wg.Done()

	// Check every minute for sources due to be fetched
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Initial fetch
	p.fetchAllSources(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.fetchAllSources(ctx)
		}
	}
}

func (p *Poller) fetchAllSources(ctx context.Context) {
	sources, err := p.store.GetActiveSources(ctx)
	if err != nil {
		p.logger.Error("failed to get active sources", "error", err)
		return
	}

	if len(sources) == 0 {
		p.logger.Debug("no active sources to fetch")
		return
	}

	var wg sync.WaitGroup
	for _, src := range sources {
		// Skip if fetched recently (within interval)
		if src.LastFetched != nil && time.Since(*src.LastFetched) < time.Duration(src.IntervalSec)*time.Second {
			continue
		}

		wg.Add(1)
		src := src
		go func() {
			defer wg.Done()
			p.fetchSource(ctx, src)
		}()
	}
	wg.Wait()
}

func (p *Poller) fetchSource(ctx context.Context, src models.Source) {
	logger := p.logger.With("source", src.Name, "feed_url", src.FeedURL)

	// Get or create circuit breaker for this source
	cb := p.getOrCreateBreaker(src.ID)

	if !cb.Allow() {
		backoff := cb.BackoffDuration()
		logger.Warn("circuit breaker open, skipping", "backoff_seconds", int(backoff.Seconds()))
		backoffUntil := time.Now().Add(backoff)
		_ = p.store.UpdateSourceHealth(ctx, src.ID, "backoff", "circuit breaker open", &backoffUntil)
		return
	}

	// Per-source timeout
	fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	feed, err := p.fetcher.Fetch(fetchCtx, src.FeedURL)
	if err != nil {
		logger.Error("fetch failed", "error", err)
		cb.Failure()
		backoff := cb.BackoffDuration()
		backoffUntil := time.Now().Add(backoff)
		_ = p.store.UpdateSourceHealth(ctx, src.ID, "backoff", err.Error(), &backoffUntil)
		return
	}

	// Parse and insert articles
	articlesInserted := 0
	for _, item := range feed.Channel.Items {
		publishedAt := parsePubDate(item.PubDate)

		// Get best image: try item.ImageURL first, then extract from content
		imgURL := item.ImageURL
		if imgURL == "" {
			imgURL = extractImageFromHTML(item.Content)
		}
		if imgURL == "" {
			imgURL = extractImageFromHTML(item.Description)
		}

		article, err := p.store.CreateArticle(ctx, models.CreateArticleInput{
			SourceID:    src.ID,
			CategoryID:  src.CategoryID,
			Title:       item.Title,
			Slug:        "",
			URL:         item.Link,
			Content:     item.Content,
			Excerpt:     truncateText(item.Description, 300),
			ImageURL:    imgURL,
			Author:      item.Author,
			PublishedAt: publishedAt,
		})
		if err == models.ErrDuplicateHash {
			continue
		}
		if err != nil {
			logger.Error("insert article failed", "error", err)
			continue
		}

		// If no image from RSS, set category default and try Wikipedia async
		if article.ImageURL == "" {
			// 1. Set category default image immediately
			if src.CategoryID != nil {
				slug := p.getCategorySlug(ctx, *src.CategoryID)
				if defaultImg := GetDefaultImage(slug); defaultImg != "" {
					p.store.UpdateArticleImage(ctx, article.ID, defaultImg)
				}
			}
			// 2. Try Wikipedia async to REPLACE default with a better image
			go p.fetchWikipediaImage(context.Background(), article.ID, item.Title)
		}

		articlesInserted++
	}

	// Success
	cb.Success()
	logger.Info("source fetched successfully", "articles", articlesInserted)
	_ = p.store.RecordFetchSuccess(ctx, src.ID, articlesInserted)
}

func (p *Poller) getOrCreateBreaker(sourceID int64) *CircuitBreaker {
	val, _ := p.breakers.LoadOrStore(sourceID, NewCircuitBreaker(3, 30*time.Second))
	return val.(*CircuitBreaker)
}

func parsePubDate(s string) time.Time {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822,
		time.RFC822Z,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	// If parsing fails, use a reasonable default
	return time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)
}

// getCategorySlug returns the slug for a category ID by scanning cached categories
func (p *Poller) getCategorySlug(ctx context.Context, categoryID int64) string {
	categories, err := p.store.ListCategories(ctx)
	if err != nil {
		return ""
	}
	for _, c := range categories {
		if c.ID == categoryID {
			return c.Slug
		}
	}
	return ""
}

// fetchWikipediaImage tries to find a Wikipedia image for an article without one
func (p *Poller) fetchWikipediaImage(ctx context.Context, articleID int64, title string) {
	keyword := ExtractKeyword(title)
	if keyword == "" {
		return
	}

	imgURL, err := p.wiki.FetchImage(ctx, keyword)
	if err != nil {
		p.logger.Debug("wikipedia image fetch failed", "keyword", keyword, "error", err)
		return
	}
	if imgURL == "" {
		return
	}

	if err := p.store.UpdateArticleImage(ctx, articleID, imgURL); err != nil {
		p.logger.Debug("update article image failed", "error", err)
	}
}

func truncateText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

