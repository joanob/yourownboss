package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware_CapturesStatusCode(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := LoggingMiddleware()
	wrappedHandler := middleware(handler)

	// Test: Handler executes and returns correct status
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestLoggingMiddleware_CapturesResponseSize(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware := LoggingMiddleware()
	wrappedHandler := middleware(handler)

	// Test: Response size is captured
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Body.Len() == 0 {
		t.Error("Expected non-empty response body")
	}
}

func TestLoggingMiddleware_HandlesMultipleWrites(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("part1"))
		w.Write([]byte("part2"))
	})

	middleware := LoggingMiddleware()
	wrappedHandler := middleware(handler)

	// Test: Multiple writes are tracked
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Body.Len() != len("part1part2") {
		t.Errorf("Expected response size 10, got %d", w.Body.Len())
	}
}

func TestLoggingMiddleware_HandlesEmptyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	middleware := LoggingMiddleware()
	wrappedHandler := middleware(handler)

	// Test: Empty response (204 No Content)
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestLoggingMiddleware_HandlesErrorResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	middleware := LoggingMiddleware()
	wrappedHandler := middleware(handler)

	// Test: Error response is handled
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestRecoveryMiddleware_Normal(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := RecoveryMiddleware()
	wrappedHandler := middleware(handler)

	// Test: Normal request without panic
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRecoveryMiddleware_PanicWithString(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("string panic")
	})

	middleware := RecoveryMiddleware()
	wrappedHandler := middleware(handler)

	// Test: String panic is recovered
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRecoveryMiddleware_PanicWithError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var x *int
		_ = *x // nil pointer dereference panic
	})

	middleware := RecoveryMiddleware()
	wrappedHandler := middleware(handler)

	// Test: Error panic is recovered
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRecoveryMiddleware_ErrorResponseFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test error")
	})

	middleware := RecoveryMiddleware()
	wrappedHandler := middleware(handler)

	// Test: Error response has correct format
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Response should be JSON
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected JSON content type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestCORSMiddleware_GET(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware("https://example.com")
	wrappedHandler := middleware(handler)

	// Test: GET request with CORS headers
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCORSMiddleware_POST(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware("https://example.com")
	wrappedHandler := middleware(handler)

	// Test: POST request with CORS headers
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCORSMiddleware_PUT(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware("https://example.com")
	wrappedHandler := middleware(handler)

	// Test: PUT request with CORS headers
	req := httptest.NewRequest("PUT", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCORSMiddleware_DELETE(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware("https://example.com")
	wrappedHandler := middleware(handler)

	// Test: DELETE request with CORS headers
	req := httptest.NewRequest("DELETE", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware("https://example.com")
	wrappedHandler := middleware(handler)

	// Test: OPTIONS preflight request returns 204
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestCORSMiddleware_Headers(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORSMiddleware("https://example.com")
	wrappedHandler := middleware(handler)

	// Test: CORS headers are set correctly
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	headers := map[string]string{
		"Access-Control-Allow-Origin":      "https://example.com",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers":     "Content-Type, Authorization",
		"Access-Control-Allow-Credentials": "true",
	}

	for key, expectedValue := range headers {
		if w.Header().Get(key) != expectedValue {
			t.Errorf("Expected %s: %s, got %s", key, expectedValue, w.Header().Get(key))
		}
	}
}
