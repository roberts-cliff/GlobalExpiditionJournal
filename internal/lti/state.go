package lti

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// StateStore manages OIDC state and nonce for LTI launches
type StateStore struct {
	mu     sync.RWMutex
	states map[string]*StateData
}

// StateData holds the state information for an LTI launch
type StateData struct {
	Nonce         string
	TargetLinkURI string
	ClientID      string
	CreatedAt     time.Time
}

// NewStateStore creates a new state store
func NewStateStore() *StateStore {
	store := &StateStore{
		states: make(map[string]*StateData),
	}
	// Start cleanup goroutine
	go store.cleanup()
	return store
}

// GenerateState creates a new state token
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateNonce creates a new nonce
func GenerateNonce() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Store saves state data
func (s *StateStore) Store(state string, data *StateData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data.CreatedAt = time.Now()
	s.states[state] = data
}

// Get retrieves and removes state data (one-time use)
func (s *StateStore) Get(state string) (*StateData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.states[state]
	if ok {
		delete(s.states, state)
	}
	return data, ok
}

// Peek retrieves state data without removing it
func (s *StateStore) Peek(state string) (*StateData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.states[state]
	return data, ok
}

// cleanup removes expired states (older than 10 minutes)
func (s *StateStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for state, data := range s.states {
			if now.Sub(data.CreatedAt) > 10*time.Minute {
				delete(s.states, state)
			}
		}
		s.mu.Unlock()
	}
}
