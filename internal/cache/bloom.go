package cache

import (
	"context"
	"hash/fnv"
	"math"

	"github.com/redis/go-redis/v9"
)

// BloomFilter implements a Bloom filter backed by Redis Bitmap.
// Used to quickly reject non-existent short codes before hitting the DB.
type BloomFilter struct {
	client    *redis.Client
	key       string
	bitSize   uint // total bits in the bitmap
	hashFuncs uint // number of hash functions
}

// NewBloomFilter creates a new Bloom filter.
// expectedItems: estimated number of items to store
// falsePositiveRate: desired false positive rate (e.g., 0.01 for 1%)
func NewBloomFilter(client *redis.Client, key string, expectedItems uint, falsePositiveRate float64) *BloomFilter {
	// Optimal bit size: m = -n*ln(p) / (ln(2)^2)
	bitSize := uint(-float64(expectedItems) * math.Log(falsePositiveRate) / (math.Ln2 * math.Ln2))
	// Optimal hash functions: k = (m/n) * ln(2)
	hashFuncs := uint(float64(bitSize) / float64(expectedItems) * math.Ln2)
	if hashFuncs < 1 {
		hashFuncs = 1
	}
	if hashFuncs > 20 {
		hashFuncs = 20
	}

	return &BloomFilter{
		client:    client,
		key:       key,
		bitSize:   bitSize,
		hashFuncs: hashFuncs,
	}
}

// Add inserts an item into the Bloom filter.
func (bf *BloomFilter) Add(ctx context.Context, item string) error {
	positions := bf.hashPositions(item)
	pipe := bf.client.Pipeline()
	for _, pos := range positions {
		pipe.SetBit(ctx, bf.key, int64(pos), 1)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// MightExist checks if an item might exist in the Bloom filter.
// Returns true if the item might exist, false if it definitely does not.
func (bf *BloomFilter) MightExist(ctx context.Context, item string) (bool, error) {
	positions := bf.hashPositions(item)
	pipe := bf.client.Pipeline()
	cmds := make([]*redis.IntCmd, len(positions))
	for i, pos := range positions {
		cmds[i] = pipe.GetBit(ctx, bf.key, int64(pos))
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	for _, cmd := range cmds {
		if cmd.Val() == 0 {
			return false, nil
		}
	}
	return true, nil
}

// LoadAll loads existing short codes into the Bloom filter from the database.
func (bf *BloomFilter) LoadAll(ctx context.Context, codes []string) error {
	for _, code := range codes {
		if err := bf.Add(ctx, code); err != nil {
			return err
		}
	}
	return nil
}

func (bf *BloomFilter) hashPositions(item string) []uint {
	h1, h2 := bf.hash(item)
	positions := make([]uint, bf.hashFuncs)
	for i := uint(0); i < bf.hashFuncs; i++ {
		pos := (h1 + i*h2) % bf.bitSize
		positions[i] = pos
	}
	return positions
}

func (bf *BloomFilter) hash(item string) (uint, uint) {
	h := fnv.New64a()
	h.Write([]byte(item))
	hashVal := h.Sum64()
	return uint(hashVal >> 32), uint(hashVal & 0xFFFFFFFF)
}
