package fsm

import (
	"sync"
	"testing"
	"time"
)

func TestMemoryStorage_SetAndGet(t *testing.T) {
	storage := NewMemoryStorage(time.Hour)
	defer storage.Stop()

	storage.Set("user1", "STATE_ONE")

	if state := storage.Get("user1"); state != "STATE_ONE" {
		t.Errorf("Expected 'STATE_ONE', got '%s'", state)
	}

	if state := storage.Get("user_unknown"); state != "" {
		t.Errorf("Expected empty string for unknown user, got '%s'", state)
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	storage := NewMemoryStorage(time.Hour)
	defer storage.Stop()

	storage.Set("user1", "STATE_ONE")
	storage.Delete("user1")

	if state := storage.Get("user1"); state != "" {
		t.Errorf("Expected empty string after deletion, got '%s'", state)
	}
}

func TestMemoryStorage_Expiration(t *testing.T) {
	// Set TTL to 100ms
	storage := NewMemoryStorage(100 * time.Millisecond)
	defer storage.Stop()

	storage.Set("user1", "STATE_ONE")

	// Verify it exists immediately
	if state := storage.Get("user1"); state != "STATE_ONE" {
		t.Errorf("Expected 'STATE_ONE' before expiration, got '%s'", state)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify it was cleaned up
	if state := storage.Get("user1"); state != "" {
		t.Errorf("Expected state to be expired and empty, got '%s'", state)
	}
}

func TestMemoryStorage_Concurrency(t *testing.T) {
	storage := NewMemoryStorage(time.Hour)
	defer storage.Stop()

	var wg sync.WaitGroup
	routines := 100

	// Launch multiple goroutines to test thread-safety (Race conditions)
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			storage.Set("user_concurrent", "STATE_BUSY")
			storage.Get("user_concurrent")
			storage.Delete("user_concurrent")
		}()
	}

	wg.Wait()
	t.Log("Concurrency test passed")
}
