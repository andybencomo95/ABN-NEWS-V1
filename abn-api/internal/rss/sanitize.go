package rss

import (
	"strings"

	"golang.org/x/net/html"
)

// allowedTags is the whitelist of HTML tags kept after sanitization.
// Anything else is dropped but its text content is preserved.
var allowedTags = map[string]bool{
	"p": true, "br": true, "hr": true, "strong": true, "b": true, "em": true,
	"i": true, "u": true, "s": true, "a": true, "ul": true, "ol": true,
	"li": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true,
	"h6": true, "img": true, "blockquote": true, "figure": true,
	"figcaption": true, "code": true, "pre": true, "span": true, "div": true,
	"table": true, "thead": true, "tbody": true, "tr": true, "td": true,
	"th": true, "sub": true, "sup": true, "small": true, "q": true,
	"cite": true, "time": true,
}

var voidTags = map[string]bool{
	"br": true, "hr": true, "img": true,
}

var allowedAttrs = map[string]bool{
	"href": true, "src": true, "alt": true, "title": true,
}

// isSafeURL blocks URL schemes that can execute code. Relative URLs and
// http(s) pass through.
func isSafeURL(u string) bool {
	lower := strings.ToLower(strings.TrimSpace(u))
	for _, p := range []string{"javascript:", "vbscript:", "data:", "file:"} {
		if strings.HasPrefix(lower, p) {
			return false
		}
	}
	return true
}

// sanitizeHTML strips dangerous tags/attributes from untrusted RSS HTML,
// keeping the text content of dropped tags. This is the single XSS defense
// point: article content is sanitized once at ingestion time.
func sanitizeHTML(input string) string {
	if !strings.Contains(input, "<") {
		return input
	}

	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		// Unparseable garbage: keep it as plain text.
		return html.EscapeString(input)
	}

	var b strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			b.WriteString(n.Data)
		case html.ElementNode:
			tag := strings.ToLower(n.Data)
			if !allowedTags[tag] {
				// Drop the tag but keep its text content.
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					walk(c)
				}
				return
			}
			b.WriteByte('<')
			b.WriteString(tag)
			for _, a := range n.Attr {
				key := strings.ToLower(a.Key)
				if !allowedAttrs[key] {
					continue
				}
				if (key == "href" || key == "src") && !isSafeURL(a.Val) {
					continue
				}
				b.WriteByte(' ')
				b.WriteString(key)
				b.WriteString(`="`)
				b.WriteString(html.EscapeString(a.Val))
				b.WriteByte('"')
			}
			b.WriteByte('>')
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
			if !voidTags[tag] {
				b.WriteString("</")
				b.WriteString(tag)
				b.WriteByte('>')
			}
		default:
			// Document/comment/doctype nodes: recurse into children only.
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}
	walk(doc)
	return b.String()
}
