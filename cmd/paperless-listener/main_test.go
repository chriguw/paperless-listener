package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestReloadHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/reload", nil)
	rr := httptest.NewRecorder()

	reloadHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestReloadHandler_LoadConfigError(t *testing.T) {
	orig := loadConfigFn
	loadConfigFn = func(string) error { return errors.New("boom") }
	t.Cleanup(func() { loadConfigFn = orig })

	req := httptest.NewRequest(http.MethodGet, "/reload", nil)
	rr := httptest.NewRecorder()

	reloadHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestReloadHandler_Success(t *testing.T) {
	orig := loadConfigFn
	loadConfigFn = func(string) error { return nil }
	t.Cleanup(func() { loadConfigFn = orig })

	req := httptest.NewRequest(http.MethodGet, "/reload", nil)
	rr := httptest.NewRecorder()

	reloadHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != "Config reloaded successfully" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func TestWebhookHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/webhook/", nil)
	rr := httptest.NewRecorder()

	webhookHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestWebhookHandler_ReadBodyError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook/", nil)
	req.Body = io.NopCloser(errReader{})
	rr := httptest.NewRecorder()

	webhookHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestWebhookHandler_InvalidDocumentIDPayload(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook/", strings.NewReader(`{"pfad":"http://example.local/no-doc/1/"}`))
	rr := httptest.NewRecorder()

	webhookHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestWebhookHandler_SuccessAcknowledgesAndProcesses(t *testing.T) {
	orig := processDocumentFn
	called := false
	gotID := ""
	processDocumentFn = func(docID string) error {
		called = true
		gotID = docID
		return nil
	}
	t.Cleanup(func() { processDocumentFn = orig })

	req := httptest.NewRequest(http.MethodPost, "/webhook/", strings.NewReader(`{"pfad":"http://paperless.local/documents/42/"}`))
	rr := httptest.NewRecorder()

	webhookHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != "Webhook received successfully" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
	if !called {
		t.Fatal("expected processDocumentFn to be called")
	}
	if gotID != "42" {
		t.Fatalf("expected docID 42, got %q", gotID)
	}
}

func TestWebhookHandler_ProcessErrorStillAcknowledges(t *testing.T) {
	orig := processDocumentFn
	processDocumentFn = func(string) error { return errors.New("update failed") }
	t.Cleanup(func() { processDocumentFn = orig })

	req := httptest.NewRequest(http.MethodPost, "/webhook/", strings.NewReader(`{"pfad":"http://paperless.local/documents/11/"}`))
	rr := httptest.NewRecorder()

	webhookHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != "Webhook received successfully" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func TestResolveConfigPath_UsesEnvOverride(t *testing.T) {
	dir := t.TempDir()
	envConfig := filepath.Join(dir, "custom-config.json")
	if err := os.WriteFile(envConfig, []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}

	t.Setenv("PAPERLESS_CONFIG_PATH", envConfig)

	got, err := resolveConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != envConfig {
		t.Fatalf("expected %q, got %q", envConfig, got)
	}
}

func TestResolveConfigPath_FallsBackToLocalConfig(t *testing.T) {
	t.Setenv("PAPERLESS_CONFIG_PATH", "")
	dir := t.TempDir()
	t.Chdir(dir)

	localConfig := filepath.Join(dir, "config.json")
	if err := os.WriteFile(localConfig, []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to create local config: %v", err)
	}

	got, err := resolveConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != configFilePath {
		t.Fatalf("expected %q, got %q", configFilePath, got)
	}
}

