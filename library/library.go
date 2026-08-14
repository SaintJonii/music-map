package library

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrNotFound is returned when a TrackRef refers to a file that no longer exists.
var ErrNotFound = errors.New("library: track not found")

// TrackRef identifies a single track within a library.
// ID is derived deterministically from the track's absolute path.
type TrackRef struct {
	ID      string
	Size    int64
	ModTime time.Time
}

// ReadSeekCloser is a readable, seekable stream that must be closed by the caller.
// It satisfies the seekable requirement of metadata.TagReader and audio.DetectFormat.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// LibrarySource discovers tracks in a library and opens streams to them.
type LibrarySource interface {
	List(ctx context.Context) ([]TrackRef, error)
	Open(ctx context.Context, ref TrackRef) (ReadSeekCloser, error)
}
