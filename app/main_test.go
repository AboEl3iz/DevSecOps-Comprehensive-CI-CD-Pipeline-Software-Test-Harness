package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	app := NewApplication(logger)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	app.Routes().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("HealthHandler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", resp.Status)
	}
}

func TestReadyHandler(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	app := NewApplication(logger)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()

	app.Routes().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ReadyHandler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test unready state
	app.isReady.Store(false)
	rr2 := httptest.NewRecorder()
	app.Routes().ServeHTTP(rr2, req)

	if status := rr2.Code; status != http.StatusServiceUnavailable {
		t.Errorf("ReadyHandler when unready returned wrong status code: got %v want %v", status, http.StatusServiceUnavailable)
	}
}

func TestDataHandler(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	app := NewApplication(logger)

	t.Run("Valid GET Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
		req.Header.Set("X-User-ID", "devsecops-tester")
		rr := httptest.NewRecorder()

		app.Routes().ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("DataHandler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var resp DataResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if resp.User != "devsecops-tester" {
			t.Errorf("Expected user 'devsecops-tester', got %v", resp.User)
		}
	})

	t.Run("Disallowed Method POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/data", nil)
		rr := httptest.NewRecorder()

		app.Routes().ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("DataHandler POST returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
		}
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	app := NewApplication(logger)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	app.Routes().ServeHTTP(rr, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
	}

	for key, expectedVal := range expectedHeaders {
		if val := rr.Header().Get(key); val != expectedVal {
			t.Errorf("Header %s = %q, want %q", key, val, expectedVal)
		}
	}
}
