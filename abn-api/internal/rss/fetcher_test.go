package rss

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestImageExtraction(t *testing.T) {
	f := NewFetcher(10 * time.Second)
	feed, err := f.Fetch(context.Background(), "https://feeds.bbci.co.uk/news/rss.xml")
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	if len(feed.Channel.Items) == 0 {
		t.Fatal("no items")
	}

	for i, item := range feed.Channel.Items {
		if i >= 3 {
			break
		}
		fmt.Printf("Item %d: MediaThumbnail=%+v, MediaContent=%+v, ImageURL=%q\n",
			i, item.MediaThumbnail, item.MediaContent, item.ImageURL)
	}

	if feed.Channel.Items[0].ImageURL == "" {
		t.Error("first item has no image URL")
	}
}
