package crawler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Options defines minimal crawl limits.
type Options struct {
	MaxPages int // maximum pages to visit
	MaxDepth int // maximum link depth from root (root = 0)
	Timeout  time.Duration
}

// Page represents extracted content from a webpage.
type Page struct {
	URL   string
	Title string
	Text  string
}

// CrawlWebsite performs a minimal breadth‑first crawl within the same host.
// It returns a slice of pages with plain text extracted from HTML bodies.
func CrawlWebsite(ctx context.Context, root string, opts Options) ([]Page, error) {
	if opts.MaxPages <= 0 {
		opts.MaxPages = 25
	}
	if opts.MaxDepth < 0 {
		opts.MaxDepth = 2
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 15 * time.Second
	}

	rootURL, err := url.Parse(root)
	if err != nil || rootURL.Scheme == "" || rootURL.Host == "" {
		return nil, errors.New("invalid root URL")
	}

	type item struct {
		u     *url.URL
		depth int
	}

	client := &http.Client{Timeout: opts.Timeout}
	queue := []item{{u: rootURL, depth: 0}}
	visited := map[string]bool{}
	pages := make([]Page, 0, 16)

	sameHost := func(u *url.URL) bool { return strings.EqualFold(u.Hostname(), rootURL.Hostname()) }

	for len(queue) > 0 && len(pages) < opts.MaxPages {
		it := queue[0]
		queue = queue[1:]

		canon := canonicalURL(it.u)
		if visited[canon] {
			continue
		}
		visited[canon] = true

		// Fetch
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, it.u.String(), nil)
		req.Header.Set("User-Agent", "VectorChatBot/1.0 (+https://example.com)")
		resp, err := client.Do(req)
		if err != nil {
			continue // skip errors
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		if ct := resp.Header.Get("Content-Type"); !strings.Contains(strings.ToLower(ct), "text/html") {
			io.Copy(io.Discard, resp.Body)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		// Parse HTML, extract text + title, and links
		doc, err := html.Parse(strings.NewReader(string(body)))
		if err != nil {
			continue
		}
		title := extractTitle(doc)
		text := extractVisibleText(doc)
		if strings.TrimSpace(text) != "" {
			pages = append(pages, Page{URL: it.u.String(), Title: title, Text: text})
		}

		// Enqueue links
		if it.depth < opts.MaxDepth {
			for _, href := range extractLinks(doc) {
				next, err := it.u.Parse(href)
				if err != nil || next == nil {
					continue
				}
				if next.Fragment != "" {
					next.Fragment = ""
				}
				if !sameHost(next) {
					continue
				}
				canonNext := canonicalURL(next)
				if visited[canonNext] {
					continue
				}
				queue = append(queue, item{u: next, depth: it.depth + 1})
				if len(queue)+len(pages) >= opts.MaxPages*2 {
					// keep queue bounded roughly
					break
				}
			}
		}
	}
	return pages, nil
}

func canonicalURL(u *url.URL) string {
	// Normalize scheme/host and strip fragment and common tracking params
	v := *u
	v.Fragment = ""
	v.Host = strings.ToLower(v.Host)
	q := v.Query()
	// remove common tracking params
	for _, k := range []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content"} {
		q.Del(k)
	}
	v.RawQuery = q.Encode()
	return v.String()
}

// extractTitle returns the content of the first <title>
func extractTitle(n *html.Node) string {
	var title string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if title != "" {
			return
		}
		if node.Type == html.ElementNode && node.Data == "title" && node.FirstChild != nil {
			title = node.FirstChild.Data
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(title)
}

// extractVisibleText collects text nodes excluding script/style/noscript.
func extractVisibleText(n *html.Node) string {
	var b strings.Builder
	var skip = map[string]bool{"script": true, "style": true, "noscript": true}
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && skip[node.Data] {
			return
		}
		if node.Type == html.TextNode {
			s := strings.TrimSpace(node.Data)
			if s != "" {
				if b.Len() > 0 {
					b.WriteString(" ")
				}
				b.WriteString(s)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(b.String())
}

// extractLinks returns hrefs from <a> tags.
func extractLinks(n *html.Node) []string {
	hrefs := make([]string, 0, 16)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, a := range node.Attr {
				if strings.EqualFold(a.Key, "href") {
					hrefs = append(hrefs, a.Val)
					break
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return hrefs
}
