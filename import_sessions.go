// ABOUTME: Tenant-owned, expiring handles for analyzed backup archives.
// ABOUTME: HTTP clients never supply filesystem paths and failed imports remain retryable.
package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"omnicollect/auth"
)

type importSession struct {
	path    string
	owner   string
	expires time.Time
	busy    bool
	timer   *time.Timer
}

type importSessions struct {
	mu      sync.Mutex
	entries map[string]*importSession
}

func importOwner(r *http.Request) string {
	owner := auth.TenantIDFromContext(r.Context())
	if owner == "" {
		return "local"
	}
	return owner
}

func (s *importSessions) register(path, owner string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.entries == nil {
		s.entries = make(map[string]*importSession)
	}
	count := 0
	for id, entry := range s.entries {
		if !entry.busy && time.Now().After(entry.expires) {
			os.Remove(entry.path)
			if entry.timer != nil {
				entry.timer.Stop()
			}
			delete(s.entries, id)
		} else if entry.owner == owner {
			count++
		}
	}
	if len(s.entries) >= 8 {
		return "", fmt.Errorf("import capacity reached; finish an existing import or wait for expiry")
	}
	if count >= 4 {
		return "", fmt.Errorf("too many pending imports; finish an existing import or wait for expiry")
	}
	id := uuid.NewString()
	entry := &importSession{path: path, owner: owner, expires: time.Now().Add(30 * time.Minute)}
	s.entries[id] = entry
	entry.timer = time.AfterFunc(30*time.Minute, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if current := s.entries[id]; current == entry && !entry.busy {
			os.Remove(entry.path)
			delete(s.entries, id)
		}
	})
	return id, nil
}

func (s *importSessions) claim(id, owner string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.entries[id]
	if entry == nil || entry.owner != owner || time.Now().After(entry.expires) {
		return "", fmt.Errorf("backup not found or expired; upload it again")
	}
	if entry.busy {
		return "", fmt.Errorf("import already in progress")
	}
	entry.busy = true
	return entry.path, nil
}

func (s *importSessions) release(id string, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.entries[id]
	if entry == nil {
		return
	}
	if success || time.Now().After(entry.expires) {
		if entry.timer != nil {
			entry.timer.Stop()
		}
		os.Remove(entry.path)
		delete(s.entries, id)
	} else {
		entry.busy = false
	}
}
