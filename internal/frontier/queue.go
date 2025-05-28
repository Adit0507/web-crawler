package frontier

import (
	"container/heap"
	"sync"
	"webcrawler/internal/models"
)

// implementin prioriity queue for urls
type PriorityQueue []*models.URL

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority // Higher priority first
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*models.URL))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// managin url to crawled with prioritization
type URLFrontier struct {
	highPriorityQueue   PriorityQueue
	mediumPriorityQueue PriorityQueue
	lowPriorityQueue    PriorityQueue
	domainQueues        map[string][]*models.URL //for politeness
	mu                  sync.Mutex
}

func NewURLFrontier() *URLFrontier {
	frontier := &URLFrontier{
		domainQueues: make(map[string][]*models.URL),
	}

	heap.Init(&frontier.highPriorityQueue)
	heap.Init(&frontier.mediumPriorityQueue)
	heap.Init(&frontier.lowPriorityQueue)

	return frontier
}

func (f *URLFrontier) AddURL(url *models.URL) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Add to priority queue based on priority
	switch url.Priority {
	case models.HighPriority:
		heap.Push(&f.highPriorityQueue, url)
	case models.MediumPriority:
		heap.Push(&f.mediumPriorityQueue, url)
	default:
		heap.Push(&f.lowPriorityQueue, url)
	}

	// Also add to domain queue for politeness
	if f.domainQueues[url.Domain] == nil {
		f.domainQueues[url.Domain] = make([]*models.URL, 0)
	}
	f.domainQueues[url.Domain] = append(f.domainQueues[url.Domain], url)
}

func (f *URLFrontier) GetNextURL() *models.URL {
	f.mu.Lock()
	defer f.mu.Unlock()

	// checkin high priority first, then medium, then low
	if f.highPriorityQueue.Len() > 0 {
		return heap.Pop(&f.highPriorityQueue).(*models.URL)
	}
	if f.mediumPriorityQueue.Len() > 0 {
		return heap.Pop(&f.mediumPriorityQueue).(*models.URL)
	}
	if f.lowPriorityQueue.Len() > 0 {
		return heap.Pop(&f.lowPriorityQueue).(*models.URL)
	}

	return nil
}

func (f *URLFrontier) IsEmpty() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.highPriorityQueue.Len() == 0 &&
		f.mediumPriorityQueue.Len() == 0 &&
		f.lowPriorityQueue.Len() == 0
} 
