package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRSSFeedCache_Singleton(t *testing.T) {
	// Get cache instance multiple times
	cache1 := GetRSSFeedCache()
	cache2 := GetRSSFeedCache()

	// Should be the same instance
	assert.Equal(t, cache1, cache2)
}

func TestCache_SetAndGet(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear() // Start fresh

	feedType := "channel"
	id := "UC123"
	params := map[string]interface{}{"limit": 10}
	data := []byte("test rss feed data")

	// Set data
	cache.Set(feedType, id, params, data)

	// Get data
	retrieved, found := cache.Get(feedType, id, params)
	assert.True(t, found)
	assert.Equal(t, data, retrieved)
}

func TestCache_GetMiss(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear()

	feedType := "channel"
	id := "UC_nonexistent"
	params := map[string]interface{}{"limit": 5}

	// Try to get non-existent data
	retrieved, found := cache.Get(feedType, id, params)
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestCache_Expiration(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear()

	feedType := "playlist"
	id := "PL123"
	params := map[string]interface{}{}
	data := []byte("expiring data")

	// Set with very short TTL
	cache.SetWithTTL(feedType, id, params, data, 50*time.Millisecond)

	// Should be found immediately
	retrieved, found := cache.Get(feedType, id, params)
	assert.True(t, found)
	assert.Equal(t, data, retrieved)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not be found after expiration
	retrieved, found = cache.Get(feedType, id, params)
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestCache_Delete(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear()

	feedType := "channel"
	id := "UC_delete"
	params := map[string]interface{}{"date": "2024-01-01"}
	data := []byte("data to delete")

	// Set data
	cache.Set(feedType, id, params, data)

	// Verify it exists
	_, found := cache.Get(feedType, id, params)
	assert.True(t, found)

	// Delete
	cache.Delete(feedType, id, params)

	// Verify it's gone
	_, found = cache.Get(feedType, id, params)
	assert.False(t, found)
}

func TestCache_Clear(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear()

	// Add multiple entries
	entries := []struct {
		feedType string
		id       string
		params   map[string]interface{}
		data     []byte
	}{
		{"channel", "UC1", map[string]interface{}{"limit": 10}, []byte("data1")},
		{"channel", "UC2", map[string]interface{}{"limit": 20}, []byte("data2")},
		{"playlist", "PL1", map[string]interface{}{}, []byte("data3")},
	}

	for _, entry := range entries {
		cache.Set(entry.feedType, entry.id, entry.params, entry.data)
	}

	// Verify all exist
	for _, entry := range entries {
		_, found := cache.Get(entry.feedType, entry.id, entry.params)
		assert.True(t, found)
	}

	// Clear cache
	cache.Clear()

	// Verify all are gone
	for _, entry := range entries {
		_, found := cache.Get(entry.feedType, entry.id, entry.params)
		assert.False(t, found)
	}
}

func TestCache_DifferentParams(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear()

	feedType := "channel"
	id := "UC_params"
	params1 := map[string]interface{}{"limit": 10}
	params2 := map[string]interface{}{"limit": 20}
	data1 := []byte("data with limit 10")
	data2 := []byte("data with limit 20")

	// Set with different params
	cache.Set(feedType, id, params1, data1)
	cache.Set(feedType, id, params2, data2)

	// Get with first params
	retrieved1, found1 := cache.Get(feedType, id, params1)
	assert.True(t, found1)
	assert.Equal(t, data1, retrieved1)

	// Get with second params
	retrieved2, found2 := cache.Get(feedType, id, params2)
	assert.True(t, found2)
	assert.Equal(t, data2, retrieved2)
}

func TestCache_SameParamsDifferentOrder(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear()

	feedType := "channel"
	id := "UC_order"

	// Note: Go's map iteration order is not guaranteed, but JSON marshaling
	// should produce the same key for the same params
	params1 := map[string]interface{}{"a": 1, "b": 2}
	params2 := map[string]interface{}{"b": 2, "a": 1}
	data := []byte("same params different order")

	// Set with first params
	cache.Set(feedType, id, params1, data)

	// Get with second params (should find it due to JSON marshaling)
	retrieved, found := cache.Get(feedType, id, params2)
	assert.True(t, found)
	assert.Equal(t, data, retrieved)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear()

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 10

	// Concurrent writes and reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			feedType := "channel"
			id := "UC_concurrent"
			params := map[string]interface{}{"goroutine": index}
			data := []byte("data from goroutine")

			for j := 0; j < numOperations; j++ {
				// Set
				cache.Set(feedType, id, params, data)

				// Get
				retrieved, found := cache.Get(feedType, id, params)
				if found {
					assert.Equal(t, data, retrieved)
				}

				// Delete occasionally
				if j%3 == 0 {
					cache.Delete(feedType, id, params)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestGenerateCacheKey_Consistency(t *testing.T) {
	tests := []struct {
		name     string
		feedType string
		id       string
		params   interface{}
		expected string
	}{
		{
			name:     "Simple params",
			feedType: "channel",
			id:       "UC123",
			params:   map[string]interface{}{"limit": 10},
			expected: "", // We'll just check consistency, not exact value
		},
		{
			name:     "No params",
			feedType: "playlist",
			id:       "PL456",
			params:   map[string]interface{}{},
			expected: "",
		},
		{
			name:     "Complex params",
			feedType: "channel",
			id:       "UC789",
			params:   map[string]interface{}{"limit": 10, "date": "2024-01-01"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate key multiple times
			key1 := generateCacheKey(tt.feedType, tt.id, tt.params)
			key2 := generateCacheKey(tt.feedType, tt.id, tt.params)

			// Should be consistent
			assert.Equal(t, key1, key2)
			assert.NotEmpty(t, key1)
		})
	}
}

func TestCache_NilParams(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear()

	feedType := "channel"
	id := "UC_nil"
	data := []byte("data with nil params")

	// Set with nil params
	cache.Set(feedType, id, nil, data)

	// Get with nil params
	retrieved, found := cache.Get(feedType, id, nil)
	assert.True(t, found)
	assert.Equal(t, data, retrieved)
}

func TestCache_LargeData(t *testing.T) {
	cache := GetRSSFeedCache()
	cache.Clear()

	feedType := "channel"
	id := "UC_large"
	params := map[string]interface{}{}
	// Create large data (~1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// Set large data
	cache.Set(feedType, id, params, largeData)

	// Get large data
	retrieved, found := cache.Get(feedType, id, params)
	assert.True(t, found)
	assert.Equal(t, largeData, retrieved)
}

// Benchmark tests
func BenchmarkCache_Set(b *testing.B) {
	cache := GetRSSFeedCache()
	feedType := "channel"
	id := "UC_bench"
	params := map[string]interface{}{"limit": 10}
	data := []byte("benchmark data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(feedType, id, params, data)
	}
}

func BenchmarkCache_Get_Hit(b *testing.B) {
	cache := GetRSSFeedCache()
	cache.Clear()
	feedType := "channel"
	id := "UC_bench_get"
	params := map[string]interface{}{"limit": 10}
	data := []byte("benchmark get data")

	cache.Set(feedType, id, params, data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(feedType, id, params)
	}
}

func BenchmarkCache_Get_Miss(b *testing.B) {
	cache := GetRSSFeedCache()
	cache.Clear()
	feedType := "channel"
	params := map[string]interface{}{"limit": 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(feedType, "UC_miss_"+string(rune(i%100)), params)
	}
}

func BenchmarkCache_Delete(b *testing.B) {
	cache := GetRSSFeedCache()
	feedType := "channel"
	params := map[string]interface{}{"limit": 10}
	data := []byte("benchmark delete data")

	// Pre-populate
	for i := 0; i < b.N; i++ {
		id := "UC_del_" + string(rune(i))
		cache.Set(feedType, id, params, data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := "UC_del_" + string(rune(i))
		cache.Delete(feedType, id, params)
	}
}

func BenchmarkGenerateCacheKey(b *testing.B) {
	feedType := "channel"
	id := "UC_key_bench"
	params := map[string]interface{}{
		"limit": 10,
		"date":  "2024-01-01",
		"extra": "param",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateCacheKey(feedType, id, params)
	}
}

func BenchmarkCache_ConcurrentReads(b *testing.B) {
	cache := GetRSSFeedCache()
	cache.Clear()
	feedType := "channel"
	id := "UC_concurrent"
	params := map[string]interface{}{"limit": 10}
	data := []byte("concurrent benchmark data")

	cache.Set(feedType, id, params, data)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get(feedType, id, params)
		}
	})
}

func BenchmarkCache_ConcurrentWrites(b *testing.B) {
	cache := GetRSSFeedCache()
	cache.Clear()
	feedType := "channel"
	id := "UC_concurrent_write"
	params := map[string]interface{}{"limit": 10}
	data := []byte("concurrent write data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Set(feedType, id, params, data)
		}
	})
}

func BenchmarkCache_MixedOperations(b *testing.B) {
	cache := GetRSSFeedCache()
	cache.Clear()
	feedType := "channel"
	params := map[string]interface{}{"limit": 10}
	data := []byte("mixed operations data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			id := "UC_mixed_" + string(rune(i%10))

			switch i % 3 {
			case 0:
				cache.Set(feedType, id, params, data)
			case 1:
				cache.Get(feedType, id, params)
			case 2:
				cache.Delete(feedType, id, params)
			}
			i++
		}
	})
}

// Test cleanup goroutine behavior
func TestCache_CleanupExpired(t *testing.T) {
	// Create a new cache instance for isolation
	testCache := &Cache{}

	feedType := "channel"
	id := "UC_cleanup"
	params := map[string]interface{}{}
	data := []byte("cleanup test data")

	// Set with short TTL
	testCache.SetWithTTL(feedType, id, params, data, 10*time.Millisecond)

	// Verify it exists
	_, found := testCache.Get(feedType, id, params)
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	// Try to get - should trigger expiration check in Get
	_, found = testCache.Get(feedType, id, params)
	assert.False(t, found)
}
