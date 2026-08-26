package rss

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type WikipediaClient struct {
	client *http.Client
}

type wikiResponse struct {
	Title     string `json:"title"`
	Thumbnail *struct {
		Source string `json:"source"`
	} `json:"thumbnail"`
}

// Wikipedia search API response
type wikiSearchResponse struct {
	Query struct {
		Search []struct {
			Title string `json:"title"`
		} `json:"search"`
	} `json:"query"`
}

// Wikimedia Commons search response
type commonsSearchResponse struct {
	Query struct {
		Search []struct {
			Title    string `json:"title"`
			ThumbURL string `json:"thumburl"`
		} `json:"search"`
	} `json:"query"`
}

func NewWikipediaClient() *WikipediaClient {
	return &WikipediaClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "is": true, "are": true,
	"was": true, "were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true, "does": true,
	"did": true, "will": true, "would": true, "could": true, "should": true,
	"may": true, "might": true, "shall": true, "can": true, "to": true,
	"of": true, "in": true, "for": true, "on": true, "with": true,
	"at": true, "by": true, "from": true, "as": true, "into": true,
	"through": true, "during": true, "before": true, "above": true,
	"below": true, "between": true, "out": true, "off": true,
	"over": true, "under": true, "again": true, "further": true,
	"then": true, "once": true, "here": true, "there": true,
	"when": true, "where": true, "why": true, "all": true,
	"each": true, "every": true, "both": true, "few": true,
	"more": true, "most": true, "other": true, "some": true,
	"such": true, "no": true, "nor": true, "not": true,
	"only": true, "own": true, "same": true, "so": true,
	"than": true, "too": true, "very": true, "just": true,
	"because": true, "but": true, "and": true, "or": true,
	"if": true, "while": true, "about": true, "up": true,
	"it": true, "its": true, "this": true, "that": true,
	"what": true, "who": true, "which": true,
	"says": true, "said": true, "report": true, "reports": true,
	"still": true, "back": true, "how": true, "their": true,
	"after": true, "gets": true, "make": true, "makes": true,
	"take": true, "takes": true, "like": true,
}

// Generic words not useful as standalone search terms
// NOTE: these are ONLY skipped when looking for multi-word phrases,
// single proper nouns like "Meta" or "OpenAI" are always tried first.
var genericWords = map[string]bool{
	"startup": true, "app": true, "ceo": true, "company": true,
	"funding": true, "raises": true, "launches": true, "acquires": true,
	"acquire": true, "acquisition": true, "partners": true, "partner": true,
	"platform": true, "service": true, "product": true,
	"build": true, "building": true, "deploy": true, "deploying": true,
	"launch": true, "hits": true, "goes": true, "gets": true,
	"bigger": true, "broader": true, "headed": true,
}

// ExtractKeywords returns multiple search candidates in priority order.
// First candidates are proper nouns (company names, product names),
// then significant common nouns. The caller should try each until one works.
func ExtractKeywords(title string) []string {
	clean := regexp.MustCompile(`[^a-zA-Z0-9\s]`).ReplaceAllString(title, " ")
	words := strings.Fields(clean)

	// Filter: keep meaningful words
	var candidates []string
	for _, w := range words {
		lower := strings.ToLower(w)
		if len(w) > 1 && !stopWords[lower] {
			candidates = append(candidates, w)
		}
		// Keep single-letter proper nouns (like "X" the company)
		if len(w) == 1 && w[0] >= 'A' && w[0] <= 'Z' {
			candidates = append(candidates, w)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	var results []string
	seen := map[string]bool{}

	// Strategy 1: Extract proper noun phrases (consecutive capitalized words)
	for i := 0; i < len(candidates); i++ {
		w := candidates[i]
		if len(w) == 0 {
			continue
		}
		// Must start with uppercase (proper noun)
		if w[0] >= 'A' && w[0] <= 'Z' {
			j := i
			phrase := []string{w}
			// Build multi-word name if next words are also capitalized
			for j+1 < len(candidates) {
				next := candidates[j+1]
				if len(next) == 0 {
					break
				}
				// Stop if next word is lowercase or generic
				if next[0] < 'A' || next[0] > 'Z' {
					break
				}
				if genericWords[strings.ToLower(next)] {
					break
				}
				j++
				phrase = append(phrase, candidates[j])
			}
			key := strings.ToLower(strings.Join(phrase, " "))
			if !seen[key] {
				seen[key] = true
				if len(phrase) > 1 {
					results = append(results, strings.Join(phrase, "_"))
				}
				results = append(results, phrase[0])
			}
			i = j // skip to end of phrase
		}
	}

	// Strategy 2: Try significant lowercase words (topic keywords)
	for _, w := range candidates {
		lower := strings.ToLower(w)
		if len(w) > 3 && !stopWords[lower] && !genericWords[lower] {
			if !seen[lower] {
				seen[lower] = true
				results = append(results, w)
			}
		}
	}

	return results
}

// ExtractKeyword is the legacy single-keyword function, kept for backward compatibility.
// Uses ExtractKeywords and returns the first candidate.
func ExtractKeyword(title string) string {
	keywords := ExtractKeywords(title)
	if len(keywords) == 0 {
		return ""
	}
	return keywords[0]
}

// FetchImage tries to find an image for the given title using multiple strategies:
// 1. Try exact Wikipedia page for each keyword candidate
// 2. Try Wikipedia search API
// 3. Try Wikimedia Commons search
func (w *WikipediaClient) FetchImage(ctx context.Context, title string) (string, error) {
	keywords := ExtractKeywords(title)
	if len(keywords) == 0 {
		return "", nil
	}

	// Strategy 1: Try exact Wikipedia page for each keyword
	for _, kw := range keywords {
		if img, err := w.fetchWikipediaPage(ctx, kw); err == nil && img != "" {
			return img, nil
		}
	}

	// Strategy 2: Wikipedia search API (best match)
	if img, err := w.searchWikipedia(ctx, title); err == nil && img != "" {
		return img, nil
	}

	// Strategy 3: Wikimedia Commons search
	if img, err := w.searchCommons(ctx, title); err == nil && img != "" {
		return img, nil
	}

	return "", nil
}

// fetchWikipediaPage tries to get the thumbnail from a specific Wikipedia page title.
func (w *WikipediaClient) fetchWikipediaPage(ctx context.Context, keyword string) (string, error) {
	if keyword == "" {
		return "", nil
	}

	apiURL := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", url.PathEscape(keyword))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ABN-News/1.0 (image-fetcher)")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	if err != nil {
		return "", err
	}

	var data wikiResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	if data.Thumbnail != nil && data.Thumbnail.Source != "" {
		return data.Thumbnail.Source, nil
	}

	return "", nil
}

// searchWikipedia uses the Wikipedia search API to find the best matching page,
// then fetches its thumbnail.
func (w *WikipediaClient) searchWikipedia(ctx context.Context, query string) (string, error) {
	// Use only the first ~6 words for search to keep it focused
	words := strings.Fields(query)
	if len(words) > 6 {
		words = words[:6]
	}
	searchQuery := strings.Join(words, " ")

	apiURL := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&srlimit=1&format=json",
		url.QueryEscape(searchQuery))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ABN-News/1.0 (image-fetcher)")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	if err != nil {
		return "", err
	}

	var data wikiSearchResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	if len(data.Query.Search) == 0 {
		return "", nil
	}

	// Fetch the thumbnail for the best matching page
	bestTitle := data.Query.Search[0].Title
	return w.fetchWikipediaPage(ctx, bestTitle)
}

// searchCommons searches Wikimedia Commons for images related to the query.
func (w *WikipediaClient) searchCommons(ctx context.Context, query string) (string, error) {
	words := strings.Fields(query)
	if len(words) > 6 {
		words = words[:6]
	}
	searchQuery := strings.Join(words, " ")

	apiURL := fmt.Sprintf("https://commons.wikimedia.org/w/api.php?action=query&list=search&srsearch=%s&srlimit=1&srnamespace=6&format=json",
		url.QueryEscape(searchQuery))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ABN-News/1.0 (image-fetcher)")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	if err != nil {
		return "", err
	}

	var data commonsSearchResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	if len(data.Query.Search) == 0 {
		return "", nil
	}

	// Get the thumbnail URL for the first result
	title := data.Query.Search[0].Title
	if data.Query.Search[0].ThumbURL != "" {
		return data.Query.Search[0].ThumbURL, nil
	}

	// Fetch thumbnail via imageinfo API
	infoURL := fmt.Sprintf("https://commons.wikimedia.org/w/api.php?action=query&titles=%s&prop=imageinfo&iiprop=url&iiurlwidth=640&format=json",
		url.QueryEscape(title))

	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, infoURL, nil)
	if err != nil {
		return "", err
	}
	req2.Header.Set("User-Agent", "ABN-News/1.0 (image-fetcher)")

	resp2, err := w.client.Do(req2)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	body2, err := io.ReadAll(io.LimitReader(resp2.Body, 100*1024))
	if err != nil {
		return "", err
	}

	var infoData struct {
		Query struct {
			Pages map[string]struct {
				ImageInfo []struct {
					ThumbURL string `json:"thumburl"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(body2, &infoData); err != nil {
		return "", err
	}

	for _, page := range infoData.Query.Pages {
		if len(page.ImageInfo) > 0 && page.ImageInfo[0].ThumbURL != "" {
			return page.ImageInfo[0].ThumbURL, nil
		}
	}

	return "", nil
}

// Default category images from Wikimedia Commons (via Wikipedia)
var categoryDefaultImages = map[string]string{
	"tecnologia": "https://upload.wikimedia.org/wikipedia/commons/thumb/7/79/Dampfturbine_Montage01.jpg/330px-Dampfturbine_Montage01.jpg",
	"general":    "https://upload.wikimedia.org/wikipedia/commons/thumb/9/9a/Land_on_the_Moon_7_21_1969-repair.jpg/330px-Land_on_the_Moon_7_21_1969-repair.jpg",
	"deportes":   "https://upload.wikimedia.org/wikipedia/commons/thumb/9/92/Youth-soccer-indiana.jpg/330px-Youth-soccer-indiana.jpg",
	"economia":   "https://upload.wikimedia.org/wikipedia/commons/thumb/6/61/Map_of_countries_by_GDP_%28PPP%29_per_capita_in_2024.svg/330px-Map_of_countries_by_GDP_%28PPP%29_per_capita_in_2024.svg.png",
	"cultura":    "https://upload.wikimedia.org/wikipedia/commons/thumb/8/8b/9_Bisonte_Magdaleniense_pol%C3%ADcromo.jpg/330px-9_Bisonte_Magdaleniense_pol%C3%ADcromo.jpg",
	"ciencia":    "https://upload.wikimedia.org/wikipedia/commons/thumb/1/1e/DNA_simple_horizontal.svg/330px-DNA_simple_horizontal.svg.png",
}

func GetDefaultImage(categorySlug string) string {
	if img, ok := categoryDefaultImages[categorySlug]; ok {
		return img
	}
	return ""
}
