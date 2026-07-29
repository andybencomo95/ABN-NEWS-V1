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
	"new": true, "first": true, "last": true, "next": true,
	"what": true, "who": true, "which": true, "get": true,
	"gets": true, "make": true, "makes": true, "take": true,
	"takes": true, "like": true, "says": true, "said": true,
	"report": true, "reports": true, "still": true, "back": true,
}

// Generic words not useful for image search
var genericWords = map[string]bool{
	"startup": true, "app": true, "ceo": true, "company": true,
	"funding": true, "raises": true, "launches": true, "acquires": true,
	"acquire": true, "acquisition": true, "partners": true, "partner": true,
	"platform": true, "service": true, "product": true, "tech": true,
	"technology": true, "software": true, "data": true, "cloud": true,
	"ai": true, "bot": true, "digital": true, "online": true,
	"global": true, "world": true, "big": true, "small": true,
	"top": true, "best": true, "build": true, "building": true,
	"deploy": true, "deploying": true, "launch": true,
}

// ExtractKeyword extracts the most significant keyword from a title
func ExtractKeyword(title string) string {
	clean := regexp.MustCompile(`[^a-zA-Z0-9\s]`).ReplaceAllString(title, " ")
	words := strings.Fields(clean)

	// Filter: keep non-stop words, skip generics
	var candidates []string
	for _, w := range words {
		lower := strings.ToLower(w)
		if len(w) > 2 && !stopWords[lower] && !genericWords[lower] {
			candidates = append(candidates, w)
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	// Find proper noun phrases (consecutive capitalized words)
	for i := 0; i < len(candidates); i++ {
		if candidates[i][0] >= 'A' && candidates[i][0] <= 'Z' && !stopWords[strings.ToLower(candidates[i])] {
			// Build multi-word name
			j := i
			for j+1 < len(candidates) && candidates[j+1][0] >= 'A' && candidates[j+1][0] <= 'Z' && !stopWords[strings.ToLower(candidates[j+1])] && !genericWords[strings.ToLower(candidates[j+1])] {
				// Skip if next word is a known generic
				if genericWords[strings.ToLower(candidates[j+1])] {
					break
				}
				j++
			}
			// Single word: return it
			if i == j {
				return candidates[i]
			}
			// Multi-word: return joined
			return strings.Join(candidates[i:j+1], "_")
		}
	}

	// Fallback: first non-stop, non-generic word
	for _, c := range candidates {
		if !genericWords[strings.ToLower(c)] {
			return c
		}
	}

	// Last resort: first candidate
	return candidates[0]
}

func (w *WikipediaClient) FetchImage(ctx context.Context, keyword string) (string, error) {
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
