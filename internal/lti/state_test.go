package lti

import (
	"testing"
)

func TestGenerateState(t *testing.T) {
	state1, err := GenerateState()
	if err != nil {
		t.Fatalf("failed to generate state: %v", err)
	}
	if state1 == "" {
		t.Error("state should not be empty")
	}

	// Should be unique
	state2, _ := GenerateState()
	if state1 == state2 {
		t.Error("states should be unique")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce1, err := GenerateNonce()
	if err != nil {
		t.Fatalf("failed to generate nonce: %v", err)
	}
	if nonce1 == "" {
		t.Error("nonce should not be empty")
	}

	// Should be unique
	nonce2, _ := GenerateNonce()
	if nonce1 == nonce2 {
		t.Error("nonces should be unique")
	}
}

func TestStateStore_StoreAndGet(t *testing.T) {
	store := &StateStore{
		states: make(map[string]*StateData),
	}

	state := "test-state"
	data := &StateData{
		Nonce:         "test-nonce",
		TargetLinkURI: "https://example.com/launch",
		ClientID:      "client-123",
	}

	store.Store(state, data)

	// Get should return and remove
	retrieved, ok := store.Get(state)
	if !ok {
		t.Fatal("expected to find state")
	}
	if retrieved.Nonce != "test-nonce" {
		t.Errorf("expected nonce 'test-nonce', got '%s'", retrieved.Nonce)
	}
	if retrieved.TargetLinkURI != "https://example.com/launch" {
		t.Errorf("expected target link URI 'https://example.com/launch', got '%s'", retrieved.TargetLinkURI)
	}

	// Should be removed after Get
	_, ok = store.Get(state)
	if ok {
		t.Error("state should have been removed after Get")
	}
}

func TestStateStore_Peek(t *testing.T) {
	store := &StateStore{
		states: make(map[string]*StateData),
	}

	state := "test-state"
	data := &StateData{
		Nonce:    "test-nonce",
		ClientID: "client-123",
	}

	store.Store(state, data)

	// Peek should return without removing
	retrieved, ok := store.Peek(state)
	if !ok {
		t.Fatal("expected to find state")
	}
	if retrieved.Nonce != "test-nonce" {
		t.Errorf("expected nonce 'test-nonce', got '%s'", retrieved.Nonce)
	}

	// Should still be there after Peek
	_, ok = store.Peek(state)
	if !ok {
		t.Error("state should still exist after Peek")
	}
}

func TestStateStore_GetNotFound(t *testing.T) {
	store := &StateStore{
		states: make(map[string]*StateData),
	}

	_, ok := store.Get("nonexistent")
	if ok {
		t.Error("should not find nonexistent state")
	}
}
