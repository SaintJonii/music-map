package metadata

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Golden HTTP fixtures as raw JSON responses from MusicBrainz API v2.

const recordingJSON = `{
  "id": "b9e9d72d-817d-4a7a-8e3a-1f4a8b0c7d2f",
  "title": "Test Recording",
  "artist-credit": [{"artist": {"id": "a1b2c3d4-1234-5678-9abc-def012345678", "name": "Test Artist"}, "name": "Test Artist"}],
  "length": 180000,
  "first-release-date": "2020-01-15",
  "isrcs": ["US-ABC-12-34567"]
}`

const isrcResultJSON = `{
  "isrc": "US-ABC-12-34567",
  "recordings": [
    {
      "id": "c8d7e6f5-4321-0abc-def0-123456789abc",
      "title": "ISRC Track",
      "artist-credit": [{"artist": {"id": "e5f6a7b8-8765-4321-abcd-ef0123456789", "name": "ISRC Artist"}, "name": "ISRC Artist"}],
      "length": 210000,
      "first-release-date": "2021-06-20"
    }
  ]
}`

// setupTestClient creates a MusicBrainzClient pointing at a test HTTP server.
func setupTestClient(handler http.HandlerFunc) (MusicBrainzClient, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := NewMusicBrainzClient(server.URL)
	return client, server
}

func TestMusicBrainzClient_LookupByMBID_Valid(t *testing.T) {
	client, server := setupTestClient(func(w http.ResponseWriter, r *http.Request) {
		// MusicBrainz API returns JSON when requested.
		if r.URL.Query().Get("fmt") == "json" || r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(recordingJSON))
	})
	defer server.Close()

	ctx := context.Background()
	result, err := client.LookupByMBID(ctx, "b9e9d72d-817d-4a7a-8e3a-1f4a8b0c7d2f")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.Title != "Test Recording" {
		t.Errorf("expected title 'Test Recording', got %q", result.Title)
	}
	if result.Artist != "Test Artist" {
		t.Errorf("expected artist 'Test Artist', got %q", result.Artist)
	}
}

func TestMusicBrainzClient_LookupByISRC(t *testing.T) {
	client, server := setupTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(isrcResultJSON))
	})
	defer server.Close()

	ctx := context.Background()
	result, err := client.LookupByISRC(ctx, "US-ABC-12-34567")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.Title != "ISRC Track" {
		t.Errorf("expected title 'ISRC Track', got %q", result.Title)
	}
	if result.Artist != "ISRC Artist" {
		t.Errorf("expected artist 'ISRC Artist', got %q", result.Artist)
	}
}

func TestMusicBrainzClient_LookupByMBID_NotFound(t *testing.T) {
	client, server := setupTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not Found"}`))
	})
	defer server.Close()

	ctx := context.Background()
	_, err := client.LookupByMBID(ctx, "00000000-0000-0000-0000-000000000000")

	if err == nil {
		t.Fatal("expected error for not-found MBID, got nil")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound via errors.Is, got: %v", err)
	}
}

func TestMusicBrainzClient_LookupByMBID_ServiceUnavailable(t *testing.T) {
	client, server := setupTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "Service temporarily unavailable"}`))
	})
	defer server.Close()

	ctx := context.Background()
	_, err := client.LookupByMBID(ctx, "b9e9d72d-817d-4a7a-8e3a-1f4a8b0c7d2f")

	if err == nil {
		t.Fatal("expected error for 503, got nil")
	}

	t.Logf("503 error message: %v", err)
}

func TestMusicBrainzClient_LookupByISRC_EmptyResult(t *testing.T) {
	client, server := setupTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"isrc": "XX-XXX-XX-00000", "recordings": []}`))
	})
	defer server.Close()

	ctx := context.Background()
	_, err := client.LookupByISRC(ctx, "XX-XXX-XX-00000")

	if err == nil {
		t.Fatal("expected ErrNotFound for ISRC with no recordings")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestMusicBrainzClient_LookupByMBID_Timeout(t *testing.T) {
	client, server := setupTestClient(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response that triggers timeout.
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	// Use a context with a short deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.LookupByMBID(ctx, "b9e9d72d-817d-4a7a-8e3a-1f4a8b0c7d2f")

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	t.Logf("timeout error: %v", err)
}

