package pokecache

import (
	"fmt"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	interval := 5 * time.Second
	cache := NewCache(interval)

	if cache == nil {
		t.Fatal("NewCache returned nil")
	}

	if cache.GetInterval() != interval {
		t.Errorf("Expected interval %v, got %v", interval, cache.GetInterval())
	}

	cacheMap := cache.GetCacheMap()
	if cacheMap == nil {
		t.Error("Cache map was not initialized")
	}

	if len(cacheMap) != 0 {
		t.Error("New cache should be empty")
	}
}

func TestCacheAdd(t *testing.T) {
	cache := NewCache(5 * time.Second)

	key := "test-key"
	value := []byte("test-value")

	cache.Add(key, value)

	cacheMap := cache.GetCacheMap()
	entry, exists := cacheMap[key]
	if !exists {
		t.Error("Added entry not found in cache")
	}

	if string(entry.Val) != string(value) {
		t.Errorf("Expected value %s, got %s", string(value), string(entry.Val))
	}

	if entry.CreatedAt.IsZero() {
		t.Error("CreatedAt timestamp was not set")
	}
}

func TestCacheGet(t *testing.T) {
	cache := NewCache(5 * time.Second)

	// Test getting existing entry
	key := "existing-key"
	expectedValue := []byte("existing-value")
	cache.Add(key, expectedValue)

	value, found := cache.Get(key)
	if !found {
		t.Error("Expected to find existing key")
	}

	if string(value) != string(expectedValue) {
		t.Errorf("Expected value %s, got %s", string(expectedValue), string(value))
	}

	// Test getting non-existing entry
	nonExistentKey := "non-existent"
	value, found = cache.Get(nonExistentKey)
	if found {
		t.Error("Expected not to find non-existent key")
	}

	if len(value) != 0 {
		t.Error("Expected empty byte slice for non-existent key")
	}
}

func TestCacheGetEmpty(t *testing.T) {
	cache := NewCache(5 * time.Second)

	// Test getting from empty cache
	value, found := cache.Get("any-key")
	if found {
		t.Error("Expected not to find key in empty cache")
	}

	if len(value) != 0 {
		t.Error("Expected empty byte slice from empty cache")
	}
}

func TestCacheMultipleEntries(t *testing.T) {
	cache := NewCache(5 * time.Second)

	entries := map[string][]byte{
		"pokemon1": []byte("pikachu"),
		"pokemon2": []byte("charizard"),
		"pokemon3": []byte("blastoise"),
	}

	// Add all entries
	for key, value := range entries {
		cache.Add(key, value)
	}

	// Verify all entries can be retrieved
	for key, expectedValue := range entries {
		value, found := cache.Get(key)
		if !found {
			t.Errorf("Expected to find key %s", key)
		}

		if string(value) != string(expectedValue) {
			t.Errorf("For key %s, expected value %s, got %s", key, string(expectedValue), string(value))
		}
	}
}

func TestCacheOverwrite(t *testing.T) {
	cache := NewCache(5 * time.Second)

	key := "overwrite-test"
	firstValue := []byte("first-value")
	secondValue := []byte("second-value")

	// Add first value
	cache.Add(key, firstValue)
	value1, found1 := cache.Get(key)
	if !found1 {
		t.Fatal("First value was not added")
	}

	// Overwrite with second value
	cache.Add(key, secondValue)
	value2, found2 := cache.Get(key)
	if !found2 {
		t.Fatal("Second value was not added")
	}

	if string(value2) != string(secondValue) {
		t.Errorf("Expected overwritten value %s, got %s", string(secondValue), string(value2))
	}

	if string(value2) == string(value1) {
		t.Error("Value was not overwritten")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := NewCache(5 * time.Second)

	// Number of concurrent goroutines
	numGoroutines := 100
	numOperationsPerGoroutine := 10

	// Channel to synchronize goroutine completion
	done := make(chan bool, numGoroutines)

	// Launch multiple goroutines that add and get from cache
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numOperationsPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d-%d", goroutineID, j)
				value := []byte(fmt.Sprintf("value-%d-%d", goroutineID, j))

				// Add to cache
				cache.Add(key, value)

				// Immediately try to get it back
				retrievedValue, found := cache.Get(key)
				if !found {
					t.Errorf("Goroutine %d: Failed to retrieve key %s", goroutineID, key)
				}

				if string(retrievedValue) != string(value) {
					t.Errorf("Goroutine %d: Expected %s, got %s", goroutineID, string(value), string(retrievedValue))
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify total number of entries
	expectedEntries := numGoroutines * numOperationsPerGoroutine
	if cache.Len() != expectedEntries {
		t.Errorf("Expected %d entries in cache, got %d", expectedEntries, cache.Len())
	}
}

func TestCacheWithNilValues(t *testing.T) {
	cache := NewCache(5 * time.Second)

	key := "nil-value-test"
	var nilValue []byte = nil

	cache.Add(key, nilValue)

	retrievedValue, found := cache.Get(key)
	if !found {
		t.Error("Expected to find entry with nil value")
	}

	if retrievedValue == nil {
		t.Error("Retrieved value should not be nil (should be empty slice)")
	}

	if len(retrievedValue) != 0 {
		t.Error("Retrieved value should be empty slice")
	}
}

func TestCacheWithEmptyValues(t *testing.T) {
	cache := NewCache(5 * time.Second)

	key := "empty-value-test"
	emptyValue := []byte{}

	cache.Add(key, emptyValue)

	retrievedValue, found := cache.Get(key)
	if !found {
		t.Error("Expected to find entry with empty value")
	}

	if len(retrievedValue) != 0 {
		t.Error("Retrieved value should be empty")
	}
}

func TestCacheEntryTimestamp(t *testing.T) {
	cache := NewCache(5 * time.Second)

	key := "timestamp-test"
	value := []byte("test-value")

	beforeAdd := time.Now()
	cache.Add(key, value)
	afterAdd := time.Now()

	cacheMap := cache.GetCacheMap()
	entry, exists := cacheMap[key]
	if !exists {
		t.Fatal("Entry not found in cache")
	}

	if entry.CreatedAt.Before(beforeAdd) || entry.CreatedAt.After(afterAdd) {
		t.Error("CreatedAt timestamp is not within expected range")
	}
}

func TestCacheExpiration(t *testing.T) {
	// Use a very short interval for testing
	interval := 200 * time.Millisecond
	cache := NewCache(interval)

	key := "expiring-key"
	value := []byte("expiring-value")

	// Add entry to cache
	cache.Add(key, value)

	// Verify it exists
	retrievedValue, found := cache.Get(key)
	if !found {
		t.Fatal("Entry should exist immediately after adding")
	}
	if string(retrievedValue) != string(value) {
		t.Fatal("Retrieved value doesn't match added value")
	}

	// Wait less than the interval - entry should still exist
	time.Sleep(interval / 2)

	// Entry should still be there
	retrievedValue, found = cache.Get(key)
	if !found {
		t.Fatal("Entry should still exist before expiration time")
	}

	// Wait for expiration and reap cycle
	time.Sleep(interval + 100*time.Millisecond)

	// Now it should be gone
	_, found = cache.Get(key)
	if found {
		t.Error("Entry should have been reaped after expiration")
	}

	// Stop the cache
	cache.Stop()
}

func TestCacheReapExpired(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	// Add some entries
	cache.Add("key1", []byte("value1"))
	time.Sleep(120 * time.Millisecond) // Wait longer than interval
	cache.Add("key2", []byte("value2"))

	// At this point, key1 should be older than interval, key2 should not
	cache.reapExpired()

	// key1 should be gone
	_, found := cache.Get("key1")
	if found {
		t.Error("key1 should have been reaped")
	}

	// key2 should still be there
	_, found = cache.Get("key2")
	if !found {
		t.Error("key2 should not have been reaped")
	}

	cache.Stop()
}

// Benchmark tests
func BenchmarkCacheAdd(b *testing.B) {
	cache := NewCache(60 * time.Second)
	value := []byte("benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i)
		cache.Add(key, value)
	}
	cache.Stop()
}

func BenchmarkCacheGet(b *testing.B) {
	cache := NewCache(60 * time.Second)

	// Pre-populate cache
	numEntries := 1000
	value := []byte("benchmark-value")
	for i := 0; i < numEntries; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i)
		cache.Add(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i%numEntries)
		cache.Get(key)
	}
	cache.Stop()
}

func BenchmarkCacheConcurrentAccess(b *testing.B) {
	cache := NewCache(60 * time.Second)
	value := []byte("concurrent-benchmark-value")

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("concurrent-key-%d", i)
			if i%2 == 0 {
				cache.Add(key, value)
			} else {
				cache.Get(key)
			}
			i++
		}
	})
	cache.Stop()
}
