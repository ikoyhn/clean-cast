package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	log "ikoyhn/podcast-sponsorblock/internal/logger"
)

const (
	DefaultTTL = 15 * time.Minute
)

type cacheEntry struct {
	data       []byte
	expiration time.Time
}

type Cache struct {
	entries sync.Map
	mu      sync.RWMutex
}

var (
	rssFeedCache *Cache
	once         sync.Once
)

// GetRSSFeedCache returns the singleton RSS feed cache instance
func GetRSSFeedCache() *Cache {
	once.Do(func() {
		rssFeedCache = &Cache{}
		// Start cleanup goroutine
		go rssFeedCache.cleanupExpired()
	})
	return rssFeedCache
}

// generateCacheKey creates a cache key from feed type, ID, and parameters
func generateCacheKey(feedType string, id string, params interface{}) string {
	// Serialize params to ensure consistent key generation
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		logger.Logger.Warn().Msgf("[CACHE] Failed to marshal params: %v", err)
		paramsJSON = []byte("{}")
	}

	// Create hash of the combination
	data := feedType + ":" + id + ":" + string(paramsJSON)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Get retrieves a value from the cache
func (c *Cache) Get(feedType string, id string, params interface{}) ([]byte, bool) {
	key := generateCacheKey(feedType, id, params)

	value, exists := c.entries.Load(key)
	if !exists {
		logger.Logger.Debug().Msgf("[CACHE] Cache miss for key: %s", key)
		return nil, false
	}

	entry := value.(*cacheEntry)

	// Check if expired
	if time.Now().After(entry.expiration) {
		logger.Logger.Debug().Msgf("[CACHE] Cache entry expired for key: %s", key)
		c.entries.Delete(key)
		return nil, false
	}

	logger.Logger.Debug().Msgf("[CACHE] Cache hit for key: %s", key)
	return entry.data, true
}

// Set stores a value in the cache with the default TTL
func (c *Cache) Set(feedType string, id string, params interface{}, data []byte) {
	c.SetWithTTL(feedType, id, params, data, DefaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL
func (c *Cache) SetWithTTL(feedType string, id string, params interface{}, data []byte, ttl time.Duration) {
	key := generateCacheKey(feedType, id, params)

	entry := &cacheEntry{
		data:       data,
		expiration: time.Now().Add(ttl),
	}

	c.entries.Store(key, entry)
	logger.Logger.Debug().Msgf("[CACHE] Stored cache entry for key: %s (TTL: %v)", key, ttl)
}

// Delete removes a value from the cache
func (c *Cache) Delete(feedType string, id string, params interface{}) {
	key := generateCacheKey(feedType, id, params)
	c.entries.Delete(key)
	logger.Logger.Debug().Msgf("[CACHE] Deleted cache entry for key: %s", key)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.entries.Range(func(key, value interface{}) bool {
		c.entries.Delete(key)
		return true
	})
	logger.Logger.Debug().Msg("[CACHE] Cleared all cache entries")
}

// cleanupExpired periodically removes expired entries from the cache
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		count := 0

		c.entries.Range(func(key, value interface{}) bool {
			entry := value.(*cacheEntry)
			if now.After(entry.expiration) {
				c.entries.Delete(key)
				count++
			}
			return true
		})

		if count > 0 {
			logger.Logger.Debug().Msgf("[CACHE] Cleaned up %d expired entries", count)
		}
	}
}
