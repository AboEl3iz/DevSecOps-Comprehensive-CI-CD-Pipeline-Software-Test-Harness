package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntegration_FullFlow(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	app := NewApplication(logger)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	client := server.Client()

	// 1. Health check
	resp, err := client.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health status %d, want %d", resp.StatusCode, http.StatusOK)
	}
	_ = resp.Body.Close()

	// 2. Readiness check
	resp, err = client.Get(server.URL + "/readyz")
	if err != nil {
		t.Fatalf("Readiness check failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ready status %d, want %d", resp.StatusCode, http.StatusOK)
	}
	_ = resp.Body.Close()

	// 3. API Data Endpoint with header
	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/data", nil)
	if err != nil {
		t.Fatalf("Failed creating request: %v", err)
	}
	req.Header.Set("X-User-ID", "integration-user")

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Data request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Data status %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed reading body: %v", err)
	}

	var dataResp DataResponse
	if err := json.Unmarshal(body, &dataResp); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	if dataResp.User != "integration-user" || dataResp.Status != "success" {
		t.Errorf("Unexpected response content: %+v", dataResp)
	}
}
