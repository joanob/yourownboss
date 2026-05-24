package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joanob/yourownboss/internal/auth/crypto"
	"github.com/joanob/yourownboss/internal/pkg/cache"
)

func TestAuthMiddleware_ValidSessionToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_SESSION_EXPIRY", "60")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_SESSION_EXPIRY")
	}()

	jwtManager, err := crypto.NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	sessionCache := cache.NewSessionCache()

	// Generate valid token
	userID := "user-123"
	sessionID := "session-456"
	companyID := "company-789"
	token, _, err := jwtManager.GenerateSessionToken(userID, sessionID, &companyID)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := r.Context().Value("user_id").(string)
		if !ok || uid == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager, sessionCache)
	wrappedHandler := middleware(handler)

	// Test: Request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: token,
	})

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_NoToken_PublicEndpoint(t *testing.T) {
	jwtManager, err := crypto.NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	sessionCache := cache.NewSessionCache()

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager, sessionCache)
	wrappedHandler := middleware(handler)

	// Test: Request without token (public endpoint)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for public endpoint without token, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtManager, err := crypto.NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	sessionCache := cache.NewSessionCache()

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager, sessionCache)
	wrappedHandler := middleware(handler)

	// Test: Request with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "invalid-token",
	})

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Should continue without auth (handler still executes)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (public handling), got %d", w.Code)
	}
}

func TestRequireAuth_WithValidAuth(t *testing.T) {
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth()
	wrappedHandler := middleware(handler)

	// Test: Request with user_id in context
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "user_id", "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireAuth_WithoutAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth()
	wrappedHandler := middleware(handler)

	// Test: Request without user_id in context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestRequireAuth_EmptyUserID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuth()
	wrappedHandler := middleware(handler)

	// Test: Request with empty user_id
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "user_id", "")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExtractFromAuthorizationHeader(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_SESSION_EXPIRY", "60")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_SESSION_EXPIRY")
	}()

	jwtManager, err := crypto.NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	sessionCache := cache.NewSessionCache()

	// Generate valid token
	token, _, err := jwtManager.GenerateSessionToken("user-123", "session-456", nil)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := r.Context().Value("user_id").(string)
		if !ok || uid == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager, sessionCache)
	wrappedHandler := middleware(handler)

	// Test: Request with Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExtractFromCookie(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret-key-at-least-32-characters-long!")
	os.Setenv("JWT_SESSION_EXPIRY", "60")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_SESSION_EXPIRY")
	}()

	jwtManager, err := crypto.NewJWTManager()
	if err != nil {
		t.Fatalf("NewJWTManager failed: %v", err)
	}

	sessionCache := cache.NewSessionCache()

	// Generate valid token
	token, _, err := jwtManager.GenerateSessionToken("user-123", "session-456", nil)
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := r.Context().Value("user_id").(string)
		if !ok || uid == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtManager, sessionCache)
	wrappedHandler := middleware(handler)

	// Test: Request with cookie
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: token,
	})

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
