package robotstxt

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type RobotsRule struct { //represents a single robots.txt rule
	UserAgent  string
	Disallow   []string
	Allow      []string
	CrawlDelay time.Duration
}

// handles robts.tx compliance
type RobotsChecker struct {
	cache     map[string]*RobotsRule
	mu        sync.RWMutex
	client    *http.Client
	userAgent string
}

func NewRobotsChecker(userAgent string) *RobotsChecker {
	return &RobotsChecker{
		cache:     make(map[string]*RobotsRule),
		userAgent: userAgent,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *RobotsChecker) CanCrawl(targetURL string) (bool, time.Duration, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return false, 0, err
	}

	domain := parsedURL.Host

	rule, err := r.getRobotsRule(domain)
	if err != nil {
		return true, 0, nil
	}

	allowed := r.isURLAllowed(parsedURL.Path, rule)

	return allowed, rule.CrawlDelay, nil
}

func (r *RobotsChecker) getRobotsRule(domain string) (*RobotsRule, error) {
	r.mu.RLock()
	if rule, exists := r.cache[domain]; exists {
		r.mu.RUnlock()
		return rule, nil
	}
	r.mu.RUnlock()

	// fetch robots.txt
	robotsURL := fmt.Sprintf("https://%s/robots.txt", domain)
	resp, err := r.client.Get(robotsURL)
	if err != nil {
		robotsURL = fmt.Sprintf("http://%s/robots.txt", domain)
		resp, err = r.client.Get(robotsURL)
		if err != nil {
			return &RobotsRule{}, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &RobotsRule{}, nil
	}

	rule := r.parseRobotsTxt(resp)

	r.mu.Lock()
	r.cache[domain] = rule
	r.mu.Unlock()

	return rule, nil
}

func (r *RobotsChecker) parseRobotsTxt(resp *http.Response) *RobotsRule {
	rule := &RobotsRule{
		UserAgent:  r.userAgent,
		Disallow:   make([]string, 0),
		Allow:      make([]string, 0),
		CrawlDelay: 0,
	}

	scanner := bufio.NewScanner(resp.Body)
	relevantSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") { //skippin empty lines and commnets
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		directive := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch directive {
		case "user-agent":
			relevantSection = (value == "*" || strings.ToLower(value) == strings.ToLower(r.userAgent))

		case "disallow":
			if relevantSection {
				rule.Disallow = append(rule.Disallow, value)
			}
		case "allow":
			if relevantSection {
				rule.Allow = append(rule.Allow, value)
			}
		case "crawl-delay":
			if relevantSection {
				if delay, err := time.ParseDuration(value + "s"); err == nil {
					rule.CrawlDelay = delay
				}
			}
		}
	}

	return rule
}

func (r *RobotsChecker) isURLAllowed(path string, rule *RobotsRule) bool {
	for _, allow := range rule.Allow {
		if allow != "" && strings.HasPrefix(path, allow) {
			return true
		}
	}

	for _, disallow := range rule.Disallow {
		if disallow == "" {
			continue
		}
		if disallow == "/" {
			return false // Disallow everything
		}
		if strings.HasPrefix(path, disallow) {
			return false
		}
	}

	return true
}
