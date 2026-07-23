package kvstore

import (
	"context"
	"bytes"
	"testing"
	"time"
)

func TestMemoryKVStore_SetGetDelete(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryKVStore(0)
	defer store.Close()

	key := "dataset-1"
	val := []byte(`{"id":"dataset-1","title":"Test Dataset"}`)

	// Set value
	if err := store.Set(ctx, key, val, 1*time.Minute); err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	// Get value
	retrieved, found, err := store.Get(ctx, key)
	if err != nil || !found {
		t.Fatalf("expected key to be found, got found=%v, err=%v", found, err)
	}

	if !bytes.Equal(retrieved, val) {
		t.Errorf("expected %s, got %s", val, retrieved)
	}

	// Delete key
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("unexpected error deleting key: %v", err)
	}

	_, found, _ = store.Get(ctx, key)
	if found {
		t.Errorf("expected key to be deleted")
	}
}

func TestMemoryKVStore_TTLExpiration(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryKVStore(10 * time.Millisecond)
	defer store.Close()

	key := "temp-key"
	val := []byte("temp-val")

	if err := store.Set(ctx, key, val, 50*time.Millisecond); err != nil {
		t.Fatalf("failed to set key: %v", err)
	}

	_, found, _ := store.Get(ctx, key)
	if !found {
		t.Fatalf("expected key to be found immediately after Set")
	}

	// Wait for TTL to expire
	time.Sleep(70 * time.Millisecond)

	_, found, _ = store.Get(ctx, key)
	if found {
		t.Errorf("expected key to expire after TTL")
	}
}

func TestMemoryKVStore_Clear(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryKVStore(0)
	defer store.Close()

	_ = store.Set(ctx, "k1", []byte("v1"), 0)
	_ = store.Set(ctx, "k2", []byte("v2"), 0)

	if err := store.Clear(ctx); err != nil {
		t.Fatalf("failed to clear store: %v", err)
	}

	_, found1, _ := store.Get(ctx, "k1")
	_, found2, _ := store.Get(ctx, "k2")

	if found1 || found2 {
		t.Errorf("expected store to be cleared, found1=%v, found2=%v", found1, found2)
	}
}
