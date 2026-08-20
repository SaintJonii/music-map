package batch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/SaintJonii/music-map/library"
	"github.com/SaintJonii/music-map/model"
)

// --- Test doubles ---

// fakeReadSeekCloser adapts an in-memory byte buffer to library.ReadSeekCloser.
type fakeReadSeekCloser struct {
	*bytes.Reader
}

func (f *fakeReadSeekCloser) Close() error { return nil }

// fakeSource implements library.LibrarySource over in-memory buffers.
type fakeSource struct {
	refs  []library.TrackRef
	files map[string][]byte
	// openDelay, when > 0, makes Open block for that long (or until ctx done).
	openDelay time.Duration
}

func newFakeSource() *fakeSource {
	return &fakeSource{files: map[string][]byte{}}
}

func (f *fakeSource) add(id string, data []byte) {
	f.refs = append(f.refs, library.TrackRef{
		ID:      id,
		Size:    int64(len(data)),
		ModTime: time.Unix(1, 0),
	})
	f.files[id] = data
}

func (f *fakeSource) List(ctx context.Context) ([]library.TrackRef, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return f.refs, nil
}

func (f *fakeSource) Open(ctx context.Context, ref library.TrackRef) (library.ReadSeekCloser, error) {
	if f.openDelay > 0 {
		select {
		case <-time.After(f.openDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	data, ok := f.files[ref.ID]
	if !ok {
		return nil, library.ErrNotFound
	}
	return &fakeReadSeekCloser{Reader: bytes.NewReader(data)}, nil
}

// fakeRepo implements Saver, recording every persisted track.
type fakeRepo struct {
	mu    sync.Mutex
	saved []model.Track
}

func (f *fakeRepo) Save(ctx context.Context, track model.Track, features model.TrackFeatures) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.saved = append(f.saved, track)
	return nil
}

func (f *fakeRepo) savedIDs() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	ids := make([]string, len(f.saved))
	for i, t := range f.saved {
		ids[i] = t.ID
	}
	return ids
}

// --- Fixture generation ---

// synthWAV builds a minimal valid 16-bit mono PCM WAV of the given sine tone.
func synthWAV(t *testing.T, sampleRate, seconds int, freq float64) []byte {
	t.Helper()

	const (
		numChannels   = 1
		bitsPerSample = 16
	)
	bytesPerSample := bitsPerSample / 8
	numSamples := sampleRate * seconds
	blockAlign := numChannels * bytesPerSample
	byteRate := sampleRate * blockAlign
	dataSize := numSamples * blockAlign

	var buf bytes.Buffer
	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16)) // fmt chunk size
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))  // PCM
	_ = binary.Write(&buf, binary.LittleEndian, uint16(numChannels))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(byteRate))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(blockAlign))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(bitsPerSample))
	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(dataSize))
	for i := 0; i < numSamples; i++ {
		sample := int16(0.5 * math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)) * 32767)
		_ = binary.Write(&buf, binary.LittleEndian, sample)
	}
	return buf.Bytes()
}

// --- Helpers ---

func validSource(t *testing.T, ids ...string) (*fakeSource, map[string][]byte) {
	t.Helper()
	src := newFakeSource()
	bytesByID := make(map[string][]byte, len(ids))
	for _, id := range ids {
		data := synthWAV(t, 8000, 1, 440)
		src.add(id, data)
		bytesByID[id] = data
	}
	return src, bytesByID
}

func setOf(ids []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return m
}

func failureIDs(fs []Failure) []string {
	ids := make([]string, len(fs))
	for i, f := range fs {
		ids[i] = f.Ref.ID
	}
	return ids
}

// --- 2.1 RED tests ---

func TestRun_ValidBatchSucceeds(t *testing.T) {
	src, _ := validSource(t, "a.wav", "b.wav", "c.wav")
	repo := &fakeRepo{}
	r := NewRunner(src, repo)

	summary, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if summary.Total != 3 {
		t.Errorf("Total = %d, want 3", summary.Total)
	}
	if summary.Succeeded != 3 {
		t.Errorf("Succeeded = %d, want 3", summary.Succeeded)
	}
	if summary.Failed != 0 {
		t.Errorf("Failed = %d, want 0 (failures: %v)", summary.Failed, summary.Failures)
	}
	if len(repo.savedIDs()) != 3 {
		t.Errorf("repo saved %d tracks, want 3", len(repo.savedIDs()))
	}
}

func TestRun_CorruptFileIsolated(t *testing.T) {
	src, _ := validSource(t, "good-a.wav", "good-b.wav")
	src.add("corrupt.mp3", []byte("this is not audio data at all, definitely corrupt"))

	repo := &fakeRepo{}
	r := NewRunner(src, repo)

	summary, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if summary.Total != 3 {
		t.Errorf("Total = %d, want 3", summary.Total)
	}
	if summary.Succeeded != 2 {
		t.Errorf("Succeeded = %d, want 2", summary.Succeeded)
	}
	if summary.Failed != 1 {
		t.Fatalf("Failed = %d, want 1 (failures: %v)", summary.Failed, summary.Failures)
	}
	if got := failureIDs(summary.Failures); len(got) != 1 || got[0] != "corrupt.mp3" {
		t.Errorf("failure IDs = %v, want [corrupt.mp3]", got)
	}

	// The valid tracks must still have been persisted.
	if got := setOf(repo.savedIDs()); !reflect.DeepEqual(got, setOf([]string{"good-a.wav", "good-b.wav"})) {
		t.Errorf("saved IDs = %v, want {good-a.wav good-b.wav}", got)
	}
}

func TestRun_MixedSummaryListsFailures(t *testing.T) {
	src, _ := validSource(t, "ok-1.wav", "ok-2.wav")
	src.add("broken-1.mp3", []byte("garbage-garbage-garbage"))
	src.add("broken-2.flac", []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05})

	repo := &fakeRepo{}
	r := NewRunner(src, repo)

	summary, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}

	if summary.Total != 4 {
		t.Errorf("Total = %d, want 4", summary.Total)
	}
	if summary.Succeeded != 2 {
		t.Errorf("Succeeded = %d, want 2", summary.Succeeded)
	}
	if summary.Failed != 2 {
		t.Fatalf("Failed = %d, want 2 (failures: %v)", summary.Failed, summary.Failures)
	}

	// Each failure must identify its file and carry a non-empty error.
	wantFailed := setOf([]string{"broken-1.mp3", "broken-2.flac"})
	gotFailed := setOf(failureIDs(summary.Failures))
	if !reflect.DeepEqual(gotFailed, wantFailed) {
		t.Errorf("failed IDs = %v, want %v", gotFailed, wantFailed)
	}
	for _, f := range summary.Failures {
		if f.Err == nil {
			t.Errorf("failure for %s has nil error", f.Ref.ID)
		}
	}
}

func TestRun_CancelMidRun(t *testing.T) {
	src := newFakeSource()
	src.openDelay = 10 * time.Millisecond
	for i := 0; i < 200; i++ {
		src.add(fmt.Sprintf("track-%03d.wav", i), synthWAV(t, 8000, 1, 440))
	}

	repo := &fakeRepo{}
	r := NewRunner(src, repo)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := r.Run(ctx)
		done <- err
	}()

	// Cancel while workers are still opening files.
	time.Sleep(25 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return promptly after cancellation")
	}
}

func TestRun_DeterministicAcrossWorkerCounts(t *testing.T) {
	// Two runs over the same batch at different worker counts must yield the
	// same success and failure SETS.
	makeSource := func() *fakeSource {
		src, _ := validSource(t, "ok-a.wav", "ok-b.wav", "ok-c.wav")
		src.add("bad-1.mp3", []byte("corrupt bytes one"))
		src.add("bad-2.mp3", []byte("corrupt bytes two"))
		return src
	}

	run := func(workers int) (success, failure map[string]struct{}) {
		src := makeSource()
		repo := &fakeRepo{}
		r := NewRunner(src, repo, WithWorkers(workers))
		summary, err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("Run(workers=%d): unexpected error: %v", workers, err)
		}
		return setOf(repo.savedIDs()), setOf(failureIDs(summary.Failures))
	}

	success1, failure1 := run(1)
	success4, failure4 := run(4)

	if !reflect.DeepEqual(success1, success4) {
		t.Errorf("success set differs across worker counts:\n  workers=1: %v\n  workers=4: %v", success1, success4)
	}
	if !reflect.DeepEqual(failure1, failure4) {
		t.Errorf("failure set differs across worker counts:\n  workers=1: %v\n  workers=4: %v", failure1, failure4)
	}
}

func TestAnalyze_FingerprintMatchesFileSHA256(t *testing.T) {
	src, bytesByID := validSource(t, "one.wav")
	repo := &fakeRepo{}
	r := NewRunner(src, repo)

	res := r.analyze(context.Background(), src.refs[0])
	if res.Err != nil {
		t.Fatalf("analyze: unexpected error: %v", res.Err)
	}

	want := sha256.Sum256(bytesByID["one.wav"])
	if res.Fingerprint != hex.EncodeToString(want[:]) {
		t.Errorf("Fingerprint = %s, want sha256 of full file %s", res.Fingerprint, hex.EncodeToString(want[:]))
	}
}
