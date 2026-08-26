package rss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:media="http://search.yahoo.com/mrss/" xmlns:content="http://purl.org/rss/1.0/modules/content/">
<channel>
  <title>Test Feed</title>
  <link>http://example.com</link>
  <description>Test feed description</description>
  <item>
    <title>First story with media content</title>
    <link>http://example.com/story1</link>
    <description>Desc 1</description>
    <media:content url="http://example.com/img1.jpg" type="image/jpeg"/>
    <content:encoded><![CDATA[<p>Story 1 body</p>]]></content:encoded>
    <pubDate>Mon, 02 Jan 2006 15:04:05 GMT</pubDate>
  </item>
  <item>
    <title>Second story with thumbnail</title>
    <link>http://example.com/story2</link>
    <description>Desc 2</description>
    <media:thumbnail url="http://example.com/thumb2.jpg" width="100" height="100"/>
  </item>
  <item>
    <title>Third story with inline img</title>
    <link>http://example.com/story3</link>
    <description><![CDATA[<p><img src="http://example.com/inline3.jpg" /></p>]]></description>
  </item>
  <item>
    <title>Fourth story no image</title>
    <link>http://example.com/story4</link>
    <description>No image here</description>
  </item>
</channel>
</rss>`

func TestFetcherImageExtraction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(testFeed))
	}))
	defer srv.Close()

	f := NewFetcher(10 * time.Second)
	feed, err := f.Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	items := feed.Channel.Items
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}

	cases := []struct {
		idx  int
		want string
	}{
		{0, "http://example.com/img1.jpg"},    // media:content priority
		{1, "http://example.com/thumb2.jpg"},  // media:thumbnail priority
		{2, "http://example.com/inline3.jpg"}, // img inside description
		{3, ""},                               // no image anywhere
	}
	for _, c := range cases {
		if got := items[c.idx].ImageURL; got != c.want {
			t.Errorf("item %d: ImageURL = %q, want %q", c.idx, got, c.want)
		}
	}

	if items[0].Content != "<p>Story 1 body</p>" {
		t.Errorf("content:encoded not parsed, got %q", items[0].Content)
	}
}

func TestExtractKeywords(t *testing.T) {
	cases := []struct {
		title    string
		contains string // one of the keywords should contain this
	}{
		{"Waymo robotaxis are headed to Munich", "Waymo"},
		{"Meta settles for $18 billion in lawsuit", "Meta"},
		{"OpenAI loses a top data center exec", "OpenAI"},
		{"Bluesky now lets you upload 10-minute long videos", "Bluesky"},
		{"Claude Cowork finally remembers what you told the app", "Claude"},
		{"SpaceX will build a second Starbase spaceport", "SpaceX"},
		{"Stability AI raises $76 million in funding", "Stability"},
		{"Instagram's First Draft feature aims to make editing Reels less tedious", "Instagram"},
		{"X sends cease-and-desist to open source project Nitter", "X"},
	}
	for _, c := range cases {
		keywords := ExtractKeywords(c.title)
		if len(keywords) == 0 {
			t.Errorf("ExtractKeywords(%q) returned no keywords", c.title)
			continue
		}
		found := false
		for _, kw := range keywords {
			if strings.Contains(strings.ToLower(kw), strings.ToLower(c.contains)) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ExtractKeywords(%q) = %v, expected to contain %q", c.title, keywords, c.contains)
		}
	}
}

func TestSanitizeHTML(t *testing.T) {
	in := `<script>alert(1)</script><p onclick="x()">Hello ` +
		`<a href="javascript:alert(1)">link</a>` +
		`<img src="http://x.com/i.jpg" onerror="alert(1)"><b>bold</b></p>`

	out := sanitizeHTML(in)

	for _, banned := range []string{"<script", "javascript:", "onerror", "onclick"} {
		if strings.Contains(out, banned) {
			t.Errorf("sanitizeHTML output still contains %q: %q", banned, out)
		}
	}
	if !strings.Contains(out, "<b>bold</b>") {
		t.Errorf("expected <b>bold</b> to survive, got %q", out)
	}
	if !strings.Contains(out, `<img src="http://x.com/i.jpg"`) {
		t.Errorf("expected safe img src to survive, got %q", out)
	}
	if !strings.Contains(out, "<a") {
		t.Errorf("expected safe <a> to survive, got %q", out)
	}
}
