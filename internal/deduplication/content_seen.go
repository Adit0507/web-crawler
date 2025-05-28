package deduplication

import (
	"crypto/md5"
	"fmt"
	"sync"
)

// trackin content hashes to detect duplicates
type ContentSeen struct {
	hashes map[string]string 
	mu     sync.RWMutex
}

func NewContentSeen() *ContentSeen {
	return &ContentSeen{
		hashes: make(map[string]string),
	}
}

func (c *ContentSeen) GetContentHash(content string) string {
	hash := md5.Sum([]byte(content))

	return fmt.Sprintf("%x", hash)
}

func (c *ContentSeen) HasSeenContent(content string) (bool, string) {
	hash := c.GetContentHash(content)
	c.mu.RLock()
	defer c.mu.RUnlock()

	if originalURL, exists := c.hashes[hash]; exists {
		return true, originalURL
	}

	return false, ""
}

func (c *ContentSeen) MarkContentSeen(content, url string) {
	hash := c.GetContentHash(content)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hashes[hash] = url
}