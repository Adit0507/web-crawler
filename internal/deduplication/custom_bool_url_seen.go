package deduplication

import "sync"

type CustomBloomURLSeen struct {
	bloomFilter   *BloomFilter
	exactSet      map[string]bool //exact set of criticla urls
	mu            sync.RWMutex
	maxExact      int
	insertedCount uint //trackin how many elements inserted
}

func NewCustomBloomURLSeen(expectedURLs uint, falsePositiveRate float64) *CustomBloomURLSeen {
	return &CustomBloomURLSeen{
		bloomFilter:   NewBlooomFilter(expectedURLs, falsePositiveRate),
		exactSet:      make(map[string]bool),
		maxExact:      1000, //keepin track of first 1000 urls
		insertedCount: 0,
	}
}

func (b *CustomBloomURLSeen) HasSeen(url string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.exactSet[url] {
		return true
	}
	// checkin bloom filter
	return b.bloomFilter.Test([]byte(url))
}

func (b *CustomBloomURLSeen) MarkSeen(url string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// addin to bloom filter
	b.bloomFilter.Add([]byte(url))
	b.insertedCount++

	// addin to exact set if not full
	if len(b.exactSet) < b.maxExact {
		b.exactSet[url] = true
	}
}

func (b *CustomBloomURLSeen) EstimatedCount() uint {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.insertedCount
}

func (b *CustomBloomURLSeen) FalsePositiveRate() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.bloomFilter.EstimateFalsePositives(b.insertedCount)
}

func (b *CustomBloomURLSeen) BloomFilterStats() (uint, uint, uint) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.bloomFilter.Size(), b.bloomFilter.HashCount(), b.insertedCount
}
