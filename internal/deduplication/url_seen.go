package deduplication

import "sync"

type URLSeen struct { //trackin urls that have been seen or crawled
	seen map[string]bool
	mu   sync.RWMutex
}

func NewURLSeen() *URLSeen {
	return &URLSeen{
		seen: make(map[string]bool),
	}
}

func (u *URLSeen) HasSeen(url string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.seen[url]
}

func (u *URLSeen) MarkSeen(url string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.seen[url] = true
}

func (u *URLSeen) Count() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return len(u.seen)
}

type URLSeenInterface interface {
	HasSeen(url string) bool
	MarkSeen(url string)
}
