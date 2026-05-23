package auth

import (
	"fmt"
	"testing"
	"time"
)

func TestSessionCache_Store_Get(t *testing.T) {
	cache := NewSessionCache()
	sessionID := "session-123"
	companyID := "company-789"
	data := SessionData{
		SessionID: sessionID,
		UserID:    "user-456",
		CompanyID: &companyID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	// Test: Store session
	cache.Store(sessionID, data)

	// Test: Get session
	retrieved := cache.Get(sessionID)
	if retrieved == nil {
		t.Fatal("Expected session to be found")
	}

	if retrieved.UserID != data.UserID {
		t.Errorf("Expected user ID %s, got %s", data.UserID, retrieved.UserID)
	}

	if retrieved.CompanyID == nil || *retrieved.CompanyID != *data.CompanyID {
		t.Errorf("Expected company ID %s, got %v", *data.CompanyID, retrieved.CompanyID)
	}
}

func TestSessionCache_Get_NotFound(t *testing.T) {
	cache := NewSessionCache()

	// Test: Get non-existent session
	retrieved := cache.Get("non-existent-session")
	if retrieved != nil {
		t.Error("Expected nil for non-existent session")
	}
}

func TestSessionCache_Get_Expired(t *testing.T) {
	cache := NewSessionCache()
	sessionID := "session-123"
	data := SessionData{
		SessionID: sessionID,
		UserID:    "user-456",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt: time.Now(),
	}

	cache.Store(sessionID, data)

	// Test: Get expired session (should return nil)
	retrieved := cache.Get(sessionID)
	if retrieved != nil {
		t.Error("Expected nil for expired session")
	}
}

func TestSessionCache_Get_Revoked(t *testing.T) {
	cache := NewSessionCache()
	sessionID := "session-123"
	now := time.Now()
	data := SessionData{
		SessionID: sessionID,
		UserID:    "user-456",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		RevokedAt: &now,
		CreatedAt: time.Now(),
	}

	cache.Store(sessionID, data)

	// Test: Get revoked session (should return nil)
	retrieved := cache.Get(sessionID)
	if retrieved != nil {
		t.Error("Expected nil for revoked session")
	}
}

func TestSessionCache_Revoke(t *testing.T) {
	cache := NewSessionCache()
	sessionID := "session-123"
	data := SessionData{
		SessionID: sessionID,
		UserID:    "user-456",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	cache.Store(sessionID, data)

	// Test: Revoke session
	cache.Revoke(sessionID)

	// Test: Get revoked session
	retrieved := cache.Get(sessionID)
	if retrieved != nil {
		t.Error("Expected nil for revoked session")
	}
}

func TestSessionCache_Delete(t *testing.T) {
	cache := NewSessionCache()
	sessionID := "session-123"
	data := SessionData{
		SessionID: sessionID,
		UserID:    "user-456",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	cache.Store(sessionID, data)

	// Test: Delete session
	cache.Delete(sessionID)

	// Test: Get deleted session
	retrieved := cache.Get(sessionID)
	if retrieved != nil {
		t.Error("Expected nil for deleted session")
	}
}

func TestSessionCache_DeleteExpired(t *testing.T) {
	cache := NewSessionCache()

	// Add multiple sessions
	cache.Store("session-1", SessionData{
		SessionID: "session-1",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt: time.Now(),
	})

	cache.Store("session-2", SessionData{
		SessionID: "session-2",
		UserID:    "user-2",
		ExpiresAt: time.Now().Add(1 * time.Hour), // Valid
		CreatedAt: time.Now(),
	})

	cache.Store("session-3", SessionData{
		SessionID: "session-3",
		UserID:    "user-3",
		ExpiresAt: time.Now().Add(-2 * time.Hour), // Expired
		CreatedAt: time.Now(),
	})

	// Test: Delete expired sessions
	count := cache.DeleteExpired()

	if count != 2 {
		t.Errorf("Expected 2 expired sessions deleted, got %d", count)
	}

	// Test: Valid session still exists
	if cache.Get("session-2") == nil {
		t.Error("Expected valid session to still exist")
	}

	// Test: Expired sessions are deleted
	if cache.Get("session-1") != nil {
		t.Error("Expected expired session-1 to be deleted")
	}

	if cache.Get("session-3") != nil {
		t.Error("Expected expired session-3 to be deleted")
	}
}

func TestSessionCache_Count(t *testing.T) {
	cache := NewSessionCache()

	// Test: Initial count
	if cache.Count() != 0 {
		t.Error("Expected initial count to be 0")
	}

	// Add sessions
	cache.Store("session-1", SessionData{
		SessionID: "session-1",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	})

	cache.Store("session-2", SessionData{
		SessionID: "session-2",
		UserID:    "user-2",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	})

	// Test: Count after adding
	if cache.Count() != 2 {
		t.Errorf("Expected count to be 2, got %d", cache.Count())
	}

	// Delete one
	cache.Delete("session-1")

	// Test: Count after deletion
	if cache.Count() != 1 {
		t.Errorf("Expected count to be 1, got %d", cache.Count())
	}
}

func TestSessionCache_Count_IgnoresExpired(t *testing.T) {
	cache := NewSessionCache()

	// Add one valid and one expired session
	cache.Store("session-1", SessionData{
		SessionID: "session-1",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(1 * time.Hour), // Valid
		CreatedAt: time.Now(),
	})

	cache.Store("session-2", SessionData{
		SessionID: "session-2",
		UserID:    "user-2",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt: time.Now(),
	})

	// Test: Count should only count valid sessions
	count := cache.Count()
	if count != 1 {
		t.Errorf("Expected count 1 (only valid), got %d", count)
	}

	// Test: Get on expired returns nil
	if cache.Get("session-2") != nil {
		t.Error("Expected Get to return nil for expired session")
	}
}

func TestSessionCache_NilCompanyID(t *testing.T) {
	cache := NewSessionCache()
	sessionID := "session-123"
	data := SessionData{
		SessionID: sessionID,
		UserID:    "user-456",
		CompanyID: nil, // User without company
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	cache.Store(sessionID, data)

	// Test: Retrieve with nil company ID
	retrieved := cache.Get(sessionID)
	if retrieved == nil {
		t.Fatal("Expected session to be found")
	}

	if retrieved.CompanyID != nil {
		t.Error("Expected nil company ID")
	}
}

func TestSessionCache_Concurrent(t *testing.T) {
	cache := NewSessionCache()
	done := make(chan bool, 100)

	// Concurrent writes
	for i := 0; i < 50; i++ {
		go func(index int) {
			sessionID := fmt.Sprintf("session-%d", index)
			userID := fmt.Sprintf("user-%d", index)
			cache.Store(sessionID, SessionData{
				SessionID: sessionID,
				UserID:    userID,
				ExpiresAt: time.Now().Add(1 * time.Hour),
				CreatedAt: time.Now(),
			})
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		go func(index int) {
			sessionID := fmt.Sprintf("session-%d", index)
			_ = cache.Get(sessionID)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Test: Final count
	if cache.Count() != 50 {
		t.Errorf("Expected 50 sessions, got %d", cache.Count())
	}
}
