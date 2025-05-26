package filters

import (
	"net/url"
	"strings"
)

type URLFilter struct {
	blockedExtensions []string
	blockedDomains    []string
	allowedDomains    []string
}

func NewURLFilter(allowedDomains []string) *URLFilter {
	return &URLFilter{
		blockedExtensions: []string{".exe", ".zip", ".pdf", ".jpg", ".png", ".gif", ".mp4", ".avi", ".mov"},
		blockedDomains:    []string{"facebook.com", "twitter.com", "instagram.com"},
		allowedDomains:    allowedDomains,
	}
}

func (f *URLFilter) ShouldCrawl(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// checkin allowed domains for focused crawling
	if len(f.allowedDomains) > 0 {
		allowed := false
		for _, domain := range f.allowedDomains {
			if strings.Contains(parsedURL.Host, domain) {
				allowed = true
				break
			}
		}

		if !allowed {
			return false
		}
	}

	// checkin blocked domains
	for _, blocked := range f.blockedDomains {
		if strings.Contains(parsedURL.Host, blocked) {
			return false
		}
	}

	path := parsedURL.Path
	for _, ext := range f.blockedExtensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return false
		}
	}

	return true
}
