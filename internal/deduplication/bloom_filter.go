package deduplication

import (
	"hash/fnv"
	"math"
)

type BloomFilter struct {
	bitArray  []bool
	size      uint
	hashCount uint
}

func NewBlooomFilter(expectedElements uint, falsePositiveRate float64) *BloomFilter {
	size := uint(-1 * float64(expectedElements) * math.Log(falsePositiveRate) / math.Log(2) * math.Log(2))

	hashCount := uint(float64(size) / float64(expectedElements) * math.Log(2))
	if hashCount == 0 {
		hashCount = 1
	}

	return &BloomFilter{
		bitArray:  make([]bool, size),
		size:      size,
		hashCount: hashCount,
	}
}

func (bf *BloomFilter) Add(data []byte) {
	hashes := bf.getHashes(data)

	for i := uint(0); i < bf.hashCount; i++ {
		index := (hashes[0] + uint64(i)*hashes[1]) % uint64(bf.size)
		bf.bitArray[index] = true
	}
}

func (bf *BloomFilter) Test(data []byte) bool {
	hashes := bf.getHashes(data)

	for i := uint(0); i < bf.hashCount; i++ {
		index := (hashes[0] + uint64(i)*hashes[1]) % uint64(bf.size)
		if !bf.bitArray[index] {
			return false
		}
	}

	return true
}

func (bf *BloomFilter) getHashes(data []byte) [2]uint64 {
	h1 := fnv.New64a()
	h1.Write(data)
	hash1 := h1.Sum64()

	h2 := fnv.New64a()
	h2.Write(data)
	h2.Write([]byte{0x42})
	hash2 := h2.Sum64()

	return [2]uint64{hash1, hash2}
}

func (bf *BloomFilter) EstimateFalsePositives(insertedElements uint) float64 {
	if insertedElements == 0 {
		return 0
	}

	 k := float64(bf.hashCount)
    n := float64(insertedElements)
    m := float64(bf.size)
    
    return math.Pow(1-math.Exp(-k*n/m), k)
}

func (bf *BloomFilter) Size() uint {
    return bf.size
}

// no. of hash functions used
func (bf *BloomFilter) HashCount() uint {
    return bf.hashCount
}
