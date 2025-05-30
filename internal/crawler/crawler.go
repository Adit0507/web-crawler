package crawler

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
	"webcrawler/internal/deduplication"
	"webcrawler/internal/downloader"
	"webcrawler/internal/filters"
	"webcrawler/internal/frontier"
	"webcrawler/internal/models"
	"webcrawler/internal/parser"
	"webcrawler/internal/robotstxt"
)

type WebCrawler struct {
	frontier       *frontier.URLFrontier
	downloader     *downloader.HTMLDownloader
	parser         *parser.ContentParser
	urlFilter      *filters.URLFilter
	urlSeen        deduplication.URLSeenInterface
	contentSeen    *deduplication.ContentSeen
	robotsChecker  *robotstxt.RobotsChecker
	maxDepth       int
	crawlDelay     time.Duration
	useBloomFilter bool
}

func NewWebCrawler(allowedDomains []string, maxDepth int, useBloomFilter bool) *WebCrawler {
	var urlSeen deduplication.URLSeenInterface

	if useBloomFilter {
		urlSeen = deduplication.NewCustomBloomURLSeen(1000000, 0.01)
	} else {
		urlSeen = deduplication.NewURLSeen()
	}

	return &WebCrawler{
		frontier:       frontier.NewURLFrontier(),
		downloader:     downloader.NewHTMLDownloader(),
		parser:         parser.NewContentParser(),
		urlFilter:      filters.NewURLFilter(allowedDomains),
		urlSeen:        urlSeen,
		contentSeen:    deduplication.NewContentSeen(),
		robotsChecker:  robotstxt.NewRobotsChecker("WebCrawler/1.0"),
		maxDepth:       maxDepth,
		crawlDelay:     1 * time.Second,
		useBloomFilter: useBloomFilter,
	}
}

func (c *WebCrawler) AddSeedURL(rawURL string, priority models.Priority) error {
	if !c.urlFilter.ShouldCrawl(rawURL) {
		return fmt.Errorf("URL filtered: %s", rawURL)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	urlObj := &models.URL{
		URL:      rawURL,
		Priority: priority,
		Depth:    0,
		Domain:   parsedURL.Host,
	}

	c.frontier.AddURL(urlObj)
	return nil
}

func (c *WebCrawler) Crawl() {
	for !c.frontier.IsEmpty() {
		urlToCrawl := c.frontier.GetNextURL()
		if urlToCrawl == nil {
			break
		}

		// checkin if already seen
		if c.urlSeen.HasSeen(urlToCrawl.URL) {
			continue
		}

		allowed, robotsDelay, err := c.robotsChecker.CanCrawl(urlToCrawl.URL)
		if err != nil {
			log.Printf("error checking robots for %s: %v", urlToCrawl.URL, err)
			continue
		}
		if !allowed {
			log.Printf("Robots.txt disallows crawling: %s", urlToCrawl.URL)
			continue
		}

		effectiveDelay := c.crawlDelay
		if robotsDelay > effectiveDelay {
			effectiveDelay = robotsDelay
		}
		// mark as seen
		c.urlSeen.MarkSeen(urlToCrawl.URL)

		if urlToCrawl.Depth > c.maxDepth { //skipin if max depth is reached
			continue
		}

		log.Printf("Crawling: %s (depth: %d, priority: %v)", urlToCrawl.URL, urlToCrawl.Depth, urlToCrawl.Priority)

		// download page
		body, err := c.downloader.Dowload(urlToCrawl.URL)
		if err != nil {
			log.Printf("Error downloading %s: %v", urlToCrawl.URL, err)
			continue
		}

		// parse content
		content, err := c.parser.ParseHTML(body, urlToCrawl.URL)
		body.Close()

		if err != nil {
			log.Printf("Error parsing %s: %v", urlToCrawl.URL, err)
			continue
		}

		// checkin for duplicate contentr
		if seen, originalURL := c.contentSeen.HasSeenContent(content.Content); seen {
			log.Printf("Duplicate content found: %s (original: %s)", urlToCrawl.URL, originalURL)
			continue
		}

		// mark content as seen
		c.contentSeen.MarkContentSeen(content.Content, urlToCrawl.URL)

		c.addDiscoveredLinks(content.Links, urlToCrawl.Depth+1, urlToCrawl.Domain)

		time.Sleep(effectiveDelay)
	}
}

func (c *WebCrawler) processContent(content *models.CrawledContent) {
	log.Printf("Processed content from %s: Title=%s, Links=%d", content.URL, content.Title, len(content.Links))
}

func (c *WebCrawler) addDiscoveredLinks(links []string, depth int, parentDomain string) {
	for _, link := range links {
		if !c.urlFilter.ShouldCrawl(link) {
			continue
		}

		if c.urlSeen.HasSeen(link) {
			continue
		}

		parsedURL, err := url.Parse(link)
		if err != nil {
			continue
		}

		priority := models.LowPriority
		if strings.Contains(parsedURL.Host, parentDomain) {
			priority = models.MediumPriority
		}

		urlObj := &models.URL{
			URL:      link,
			Priority: priority,
			Depth:    depth,
			Domain:   parsedURL.Host,
		}

		c.frontier.AddURL(urlObj)
	}
}
