package library

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Scanner is a local-filesystem LibrarySource that walks a folder recursively
// and discovers audio tracks by extension (.wav, .mp3, .flac).
type Scanner struct {
	root string
}

// NewScanner creates a Scanner rooted at the given path.
func NewScanner(root string) *Scanner {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	return &Scanner{root: abs}
}

// List walks the scanner's root and returns a TrackRef for every audio file.
// Directories are skipped; non-audio files are ignored. An unreadable folder
// aborts the walk with a descriptive error.
func (s *Scanner) List(ctx context.Context) ([]TrackRef, error) {
	var refs []TrackRef

	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !isAudioExt(filepath.Ext(path)) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		refs = append(refs, TrackRef{
			ID:      path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("library: scan %s: %w", s.root, err)
	}

	if refs == nil {
		refs = []TrackRef{}
	}
	return refs, nil
}

// Open opens the file identified by ref and returns it as a ReadSeekCloser.
// It returns ErrNotFound when the file no longer exists.
func (s *Scanner) Open(ctx context.Context, ref TrackRef) (ReadSeekCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	f, err := os.Open(ref.ID)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("library: open %s: %w", ref.ID, ErrNotFound)
		}
		return nil, fmt.Errorf("library: open %s: %w", ref.ID, err)
	}
	return f, nil
}

// isAudioExt reports whether ext is a supported audio extension (case-insensitive).
func isAudioExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".wav", ".mp3", ".flac":
		return true
	default:
		return false
	}
}
