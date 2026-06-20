// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package chwriter

import (
	"sync"
	"time"
)

const dedupTTL = 2 * time.Minute

type dedupSet struct {
	mu          sync.Mutex
	keys        map[string]time.Time
	lastCleanup time.Time
}

func newDedupSet() *dedupSet {
	return &dedupSet{
		keys:        make(map[string]time.Time),
		lastCleanup: time.Now(),
	}
}

// markIfNew records key when it has not been seen within dedupTTL.
func (s *dedupSet) markIfNew(key string) bool {
	if key == "" {
		return false
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	// Periodically clean up all expired keys (e.g., every 30 seconds)
	if now.Sub(s.lastCleanup) >= 30*time.Second {
		for existing, expiresAt := range s.keys {
			if now.After(expiresAt) {
				delete(s.keys, existing)
			}
		}
		s.lastCleanup = now
	}

	if expiresAt, exists := s.keys[key]; exists && now.Before(expiresAt) {
		return false
	}
	s.keys[key] = now.Add(dedupTTL)
	return true
}