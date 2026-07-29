package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	httpClient *http.Client
	apiURL     string
	cache      sync.Map // key -> string (translated text)
}

type myMemoryResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
		Match          float64 `json:"match"`
	} `json:"responseData"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiURL: "https://api.mymemory.translated.net/get",
	}
}

func (c *Client) Translate(ctx context.Context, text string, sourceLang, targetLang string) (string, error) {
	if text == "" || strings.TrimSpace(text) == "" {
		return text, nil
	}

	// Check in-memory cache
	cacheKey := fmt.Sprintf("%s:%s:%s", sourceLang, targetLang, text)
	if val, ok := c.cache.Load(cacheKey); ok {
		return val.(string), nil
	}

	// Skip short texts (likely names, brands, or dates)
	if len([]rune(text)) < 5 {
		return text, nil
	}

	// Call MyMemory API
	params := url.Values{}
	params.Set("q", text)
	params.Set("langpair", fmt.Sprintf("%s|%s", sourceLang, targetLang))

	reqURL := fmt.Sprintf("%s?%s", c.apiURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return text, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "ABN-News/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return text, fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	if err != nil {
		return text, fmt.Errorf("read response: %w", err)
	}

	var result myMemoryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return text, fmt.Errorf("parse response: %w", err)
	}

	translated := result.ResponseData.TranslatedText
	if translated == "" {
		return text, nil
	}

	// Cache in memory
	c.cache.Store(cacheKey, translated)

	return translated, nil
}

func (c *Client) TranslateBatch(ctx context.Context, texts []string, sourceLang, targetLang string) ([]string, error) {
	results := make([]string, len(texts))

	// Translate sequentially to avoid rate limits
	for i, text := range texts {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		translated, err := c.Translate(ctx, text, sourceLang, targetLang)
		if err != nil {
			// Log error but continue with original text
			results[i] = text
			continue
		}
		results[i] = translated

		// Small delay to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	return results, nil
}
