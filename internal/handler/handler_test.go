package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/course/backend-go/internal/handler"
)

// TestHealthz memverifikasi health check endpoint.
// Tidak ada NR instrumentasi di Healthz — tes sederhana.
func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handler.Healthz(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %s", body["status"])
	}
}

// TestHello memverifikasi Hello endpoint dengan NR agent tidak aktif.
// Ketika NEW_RELIC_LICENSE_KEY kosong, semua NR calls di-skip (nil-safe).
// Test ini membuktikan handler berjalan normal TANPA NR agent (graceful degradation).
func TestHello(t *testing.T) {
	// Tidak perlu set NEW_RELIC_LICENSE_KEY — handler gracefully handle nil txn
	req := httptest.NewRequest(http.MethodGet, "/api/v1/hello", nil)
	w := httptest.NewRecorder()

	handler.Hello(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["message"] == "" {
		t.Fatal("expected non-empty message")
	}
}

// TestVersion memverifikasi Version endpoint dengan APP_VERSION env var.
func TestVersion(t *testing.T) {
	t.Setenv("APP_VERSION", "test-123")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	w := httptest.NewRecorder()

	handler.Version(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["version"] != "test-123" {
		t.Fatalf("expected version=test-123, got %s", body["version"])
	}
}
