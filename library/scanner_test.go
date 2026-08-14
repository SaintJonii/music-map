package library

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile creates a file at path with the given content, creating parent dirs.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestScanner_List_Scenarios(t *testing.T) {
	t.Run("nested scan", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "a.wav"), "wav-root")
		writeFile(t, filepath.Join(root, "sub", "b.mp3"), "mp3-nested")
		writeFile(t, filepath.Join(root, "sub", "deep", "c.flac"), "flac-deep")

		s := NewScanner(root)
		refs, err := s.List(context.Background())
		if err != nil {
			t.Fatalf("List: unexpected error: %v", err)
		}
		if len(refs) != 3 {
			t.Fatalf("expected 3 tracks, got %d: %v", len(refs), refs)
		}
		ids := refIDs(refs)
		for _, want := range []string{"a.wav", "b.mp3", "c.flac"} {
			if !containsSuffix(ids, want) {
				t.Errorf("missing track ending in %s: %v", want, ids)
			}
		}
		for _, r := range refs {
			if !filepath.IsAbs(r.ID) {
				t.Errorf("expected absolute ID, got %q", r.ID)
			}
		}
	})

	t.Run("non-audio ignored", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "song.wav"), "wav")
		writeFile(t, filepath.Join(root, "cover.jpg"), "jpg")
		writeFile(t, filepath.Join(root, "notes.txt"), "txt")
		writeFile(t, filepath.Join(root, "data.bin"), "bin")

		s := NewScanner(root)
		refs, err := s.List(context.Background())
		if err != nil {
			t.Fatalf("List: unexpected error: %v", err)
		}
		if len(refs) != 1 {
			t.Fatalf("expected 1 audio track, got %d: %v", len(refs), refs)
		}
		if !strings.HasSuffix(refs[0].ID, "song.wav") {
			t.Errorf("expected song.wav, got %s", refs[0].ID)
		}
	})

	t.Run("case-insensitive extensions", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "A.WAV"), "wav")
		writeFile(t, filepath.Join(root, "B.Mp3"), "mp3")
		writeFile(t, filepath.Join(root, "C.Flac"), "flac")

		s := NewScanner(root)
		refs, err := s.List(context.Background())
		if err != nil {
			t.Fatalf("List: unexpected error: %v", err)
		}
		if len(refs) != 3 {
			t.Fatalf("expected 3 tracks (case-insensitive), got %d: %v", len(refs), refs)
		}
	})

	t.Run("size and modtime populated", func(t *testing.T) {
		root := t.TempDir()
		p := filepath.Join(root, "song.wav")
		writeFile(t, p, "12345")

		s := NewScanner(root)
		refs, err := s.List(context.Background())
		if err != nil {
			t.Fatalf("List: unexpected error: %v", err)
		}
		if len(refs) != 1 {
			t.Fatalf("expected 1 track, got %d", len(refs))
		}
		if refs[0].Size != 5 {
			t.Errorf("expected size 5, got %d", refs[0].Size)
		}
		if refs[0].ModTime.IsZero() {
			t.Error("expected non-zero ModTime")
		}
	})

	t.Run("empty folder", func(t *testing.T) {
		root := t.TempDir()
		s := NewScanner(root)
		refs, err := s.List(context.Background())
		if err != nil {
			t.Fatalf("List: expected no error for empty folder, got %v", err)
		}
		if len(refs) != 0 {
			t.Fatalf("expected empty list, got %v", refs)
		}
	})

	t.Run("unreadable folder", func(t *testing.T) {
		root := t.TempDir()
		locked := filepath.Join(root, "locked")
		writeFile(t, filepath.Join(locked, "a.wav"), "wav")
		if err := os.Chmod(locked, 0o000); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

		// If the OS still allows reads (e.g. running as root), we cannot
		// simulate an unreadable directory — skip rather than assert falsely.
		if _, err := os.ReadDir(locked); err == nil {
			t.Skip("running privileged; cannot simulate unreadable directory")
		}

		s := NewScanner(root)
		_, err := s.List(context.Background())
		if err == nil {
			t.Fatal("List: expected error for unreadable folder, got nil")
		}
	})
}

func TestScanner_Open_Scenarios(t *testing.T) {
	t.Run("open listed track", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "song.wav"), "RIFFhello-audio-bytes")

		s := NewScanner(root)
		refs, err := s.List(context.Background())
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(refs) != 1 {
			t.Fatalf("expected 1 track, got %d", len(refs))
		}

		rsc, err := s.Open(context.Background(), refs[0])
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		defer rsc.Close()

		data, err := io.ReadAll(rsc)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		if string(data) != "RIFFhello-audio-bytes" {
			t.Errorf("expected stream content, got %q", string(data))
		}

		// Prove seekability: rewind to the start.
		if _, err := rsc.Seek(0, io.SeekStart); err != nil {
			t.Fatalf("Seek: %v", err)
		}
	})

	t.Run("missing file on open", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "song.mp3"), "mp3")

		s := NewScanner(root)
		refs, err := s.List(context.Background())
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(refs) != 1 {
			t.Fatalf("expected 1 track, got %d", len(refs))
		}

		if err := os.Remove(refs[0].ID); err != nil {
			t.Fatalf("remove: %v", err)
		}

		_, err = s.Open(context.Background(), refs[0])
		if err == nil {
			t.Fatal("Open: expected error for missing file, got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Open: expected ErrNotFound, got %v", err)
		}
	})
}

// refIDs returns the basenames of the given refs' IDs for stable comparison.
func refIDs(refs []TrackRef) []string {
	ids := make([]string, len(refs))
	for i, r := range refs {
		ids[i] = filepath.Base(r.ID)
	}
	return ids
}

func containsSuffix(paths []string, suffix string) bool {
	for _, p := range paths {
		if strings.HasSuffix(p, suffix) {
			return true
		}
	}
	return false
}
