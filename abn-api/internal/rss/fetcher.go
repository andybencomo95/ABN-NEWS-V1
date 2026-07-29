package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Feed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Link        string `xml:"link"`
		Items       []Item `xml:"item"`
	} `xml:"channel"`
}

type MediaObject struct {
	URL    string `xml:"url,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
	Type   string `xml:"type,attr"`
}

type Enclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length int    `xml:"length,attr"`
}

type Item struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Content     string   `xml:"encoded"`
	PubDate     string   `xml:"pubDate"`
	Author      string   `xml:"author"`
	GUID        string   `xml:"guid"`
	Categories  []string `xml:"category"`
	Enclosure   *Enclosure `xml:"enclosure"`

	// Media RSS (after namespace replacement)
	MediaContent    *MediaObject `xml:"media_content"`
	MediaThumbnail  *MediaObject `xml:"media_thumbnail"`

	// Extracted image URL
	ImageURL string `xml:"-"`
}

var imgSrcRegex = regexp.MustCompile(`<img[^>]+src="([^"]+)"`)

func extractImageFromHTML(html string) string {
	if html == "" {
		return ""
	}
	matches := imgSrcRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

type Fetcher struct {
	client *http.Client
}

func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
			},
		},
	}
}

func (f *Fetcher) Fetch(ctx context.Context, feedURL string) (*Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "ABN-News-RSS/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	// Pre-process: rename namespaced elements so Go XML can parse them
	processed := strings.NewReplacer(
		"<media:content", "<media_content",
		"</media:content", "</media_content",
		"<media:thumbnail", "<media_thumbnail",
		"</media:thumbnail", "</media_thumbnail",
		"<content:encoded", "<encoded",
		"</content:encoded", "</encoded",
	).Replace(string(body))

	var feed Feed
	if err := xml.Unmarshal([]byte(processed), &feed); err != nil {
		return nil, fmt.Errorf("parse xml: %w", err)
	}

	// Post-process: extract best image URL for each item
	for i := range feed.Channel.Items {
		item := &feed.Channel.Items[i]

		// Priority 1: media:content (NYT, etc)
		if item.MediaContent != nil && item.MediaContent.URL != "" {
			item.ImageURL = item.MediaContent.URL
			continue
		}

		// Priority 2: media:thumbnail (BBC, etc)
		if item.MediaThumbnail != nil && item.MediaThumbnail.URL != "" {
			item.ImageURL = item.MediaThumbnail.URL
			continue
		}

		// Priority 3: enclosure with image type
		if item.Enclosure != nil && item.Enclosure.URL != "" &&
			strings.HasPrefix(item.Enclosure.Type, "image/") {
			item.ImageURL = item.Enclosure.URL
			continue
		}

		// Priority 4: First img in content HTML
		if img := extractImageFromHTML(item.Content); img != "" {
			item.ImageURL = img
			continue
		}

		// Priority 5: First img in description
		if img := extractImageFromHTML(item.Description); img != "" {
			item.ImageURL = img
			continue
		}
	}

	return &feed, nil
}
