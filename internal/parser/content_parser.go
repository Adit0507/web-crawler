package parser

import (
	"io"
	"net/url"
	"strings"
	"webcrawler/internal/models"

	"golang.org/x/net/html"
)

// extracts links and content from html
type ContentParser struct{}

func NewContentParser() *ContentParser {
	return &ContentParser{}
}

func (p *ContentParser) ParseHTML(body io.Reader, baseURL string) (*models.CrawledContent, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	content := &models.CrawledContent{
		URL:   baseURL,
		Links: make([]string, 0),
	}

	p.extractContent(doc, content, baseURL)

	return content, nil
}

func (p *ContentParser) extractContent(n *html.Node, content *models.CrawledContent, baseURL string) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "title":
			if n.FirstChild != nil {
				content.Title = strings.TrimSpace(n.FirstChild.Data)
			}

		case "a":
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					if link := p.resolveURL(attr.Val, baseURL); link != "" {
						content.Links = append(content.Links, link)
					}
				}
			}
		case "p", "div", "span", "h1", "h2", "h3", "h4", "h5", "h6":
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				text := strings.TrimSpace(n.FirstChild.Data)
				if text != "" {
					content.Content += text + " "
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.extractContent(c, content, baseURL)
	}

}

func (p *ContentParser) resolveURL(href, baseURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(ref)

	return resolved.String()
}
