// Package batch runs concurrent analysis of a music library.
//
// Workers only compute; a single collector goroutine owns every persistence
// write, which keeps the storage layer single-writer by construction.
package batch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/SaintJonii/music-map/audio"
	"github.com/SaintJonii/music-map/library"
	"github.com/SaintJonii/music-map/metadata"
	"github.com/SaintJonii/music-map/model"
)

// AnalyzedTrack is the product of analyzing a single track, carrying the
// storage-level dedupe metadata (fingerprint, size, mod_time). Those fields
// stay out of model.Track, mirroring the storage layer's contract.
type AnalyzedTrack struct {
	Track       model.Track
	Features    model.TrackFeatures
	Fingerprint string
	Size        int64
	ModTime     time.Time
}

// Saver persists an analyzed track, applying the dedupe policy. It reports
// whether the track was skipped (already analyzed / unchanged) rather than
// saved. A non-nil error is a genuine persistence failure.
type Saver interface {
	SaveAnalyzed(ctx context.Context, a AnalyzedTrack) (skipped bool, err error)
}

// Result is the outcome of analyzing a single track.
// A non-nil Err marks the track as failed; otherwise Track, Features and
// Fingerprint carry the analysis products.
type Result struct {
	Ref         library.TrackRef
	Track       model.Track
	Features    model.TrackFeatures
	Fingerprint string
	Err         error
}

// Failure records a track that could not be analyzed or persisted.
type Failure struct {
	Ref library.TrackRef
	Err error
}

// Summary aggregates the outcome of a Run.
type Summary struct {
	Total     int
	Succeeded int
	Skipped   int
	Failed    int
	Failures  []Failure
}

// Runner analyzes every track of a LibrarySource concurrently and persists
// results through a Saver.
type Runner struct {
	src       library.LibrarySource
	repo      Saver
	tagReader metadata.TagReader
	extractor audio.FeatureExtractor
	workers   int
}

// Option configures a Runner.
type Option func(*Runner)

// WithWorkers overrides the number of worker goroutines. The default is
// runtime.GOMAXPROCS(0).
func WithWorkers(n int) Option {
	return func(r *Runner) { r.workers = n }
}

// NewRunner builds a Runner over src, persisting through repo.
func NewRunner(src library.LibrarySource, repo Saver, opts ...Option) *Runner {
	r := &Runner{
		src:       src,
		repo:      repo,
		tagReader: metadata.NewTagReader(),
		extractor: &audio.DefaultExtractor{},
		workers:   runtime.GOMAXPROCS(0),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Run lists the source and analyzes every track concurrently, aggregating a
// Summary. A single collector goroutine (this function's own loop) owns every
// repo.Save call.
//
// On cancellation Run returns promptly with the partial summary accumulated so
// far and the context error.
func (r *Runner) Run(ctx context.Context) (Summary, error) {
	refs, err := r.src.List(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("batch: list: %w", err)
	}

	jobs := make(chan library.TrackRef)
	results := make(chan Result)

	workers := r.workers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go r.worker(ctx, jobs, results, &wg)
	}

	// Feeder publishes refs to workers, stopping early on cancellation.
	go func() {
		defer close(jobs)
		for _, ref := range refs {
			select {
			case jobs <- ref:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Closer closes results once every worker has finished.
	go func() {
		wg.Wait()
		close(results)
	}()

	summary := Summary{Total: len(refs)}
	for {
		select {
		case res, ok := <-results:
			if !ok {
				return summary, nil
			}
			if res.Err != nil {
				summary.Failed++
				summary.Failures = append(summary.Failures, Failure{Ref: res.Ref, Err: res.Err})
				continue
			}
			skipped, err := r.repo.SaveAnalyzed(ctx, AnalyzedTrack{
				Track:       res.Track,
				Features:    res.Features,
				Fingerprint: res.Fingerprint,
				Size:        res.Ref.Size,
				ModTime:     res.Ref.ModTime,
			})
			if err != nil {
				summary.Failed++
				summary.Failures = append(summary.Failures, Failure{
					Ref: res.Ref,
					Err: fmt.Errorf("save: %w", err),
				})
				continue
			}
			if skipped {
				summary.Skipped++
			} else {
				summary.Succeeded++
			}
		case <-ctx.Done():
			return summary, ctx.Err()
		}
	}
}

// worker drains jobs, analyzes each track, and forwards the result to the
// collector. Sends are guarded by the context so a cancelled run never leaves
// a worker blocked on an unread channel.
func (r *Runner) worker(ctx context.Context, jobs <-chan library.TrackRef, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for ref := range jobs {
		res := r.analyze(ctx, ref)
		select {
		case results <- res:
		case <-ctx.Done():
			return
		}
	}
}

// analyze runs the single-file pipeline on one track: open → read tags →
// detect format → decode (tee-ing bytes into a SHA-256 hash) → extract
// features. The resulting fingerprint is carried in Result for the storage
// layer to persist in a later change.
func (r *Runner) analyze(ctx context.Context, ref library.TrackRef) Result {
	if err := ctx.Err(); err != nil {
		return Result{Ref: ref, Err: err}
	}

	f, err := r.src.Open(ctx, ref)
	if err != nil {
		return Result{Ref: ref, Err: fmt.Errorf("open: %w", err)}
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()

	// Tags first (existing strategy); the tag reader only needs the header.
	track, err := r.tagReader.ReadTags(f)
	if err != nil {
		return Result{Ref: ref, Err: fmt.Errorf("read tags: %w", err)}
	}
	track.FilePath = ref.ID
	if track.ID == "" {
		track.ID = ref.ID
	}

	// Rewind so the hash covers the whole file from the start.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return Result{Ref: ref, Err: fmt.Errorf("seek: %w", err)}
	}

	decoder, rd, err := audio.DetectFormat(io.TeeReader(f, h))
	if err != nil {
		return Result{Ref: ref, Err: fmt.Errorf("detect format: %w", err)}
	}
	samples, sampleRate, channels, err := decoder.Decode(rd)
	if err != nil {
		return Result{Ref: ref, Err: fmt.Errorf("decode: %w", err)}
	}
	if channels <= 0 {
		channels = 1
	}
	track.Duration = float64(len(samples)) / float64(sampleRate*channels)

	features := r.extract(samples, sampleRate)

	return Result{
		Ref:         ref,
		Track:       track,
		Features:    features,
		Fingerprint: hex.EncodeToString(h.Sum(nil)),
	}
}

// extract derives the acoustic features from decoded samples, mirroring the
// single-file pipeline. Feature extraction errors are non-fatal: the failed
// feature contributes its zero value, matching the existing CLI behavior.
func (r *Runner) extract(samples []float64, sampleRate int) model.TrackFeatures {
	energy := r.extractor.ExtractRMS(samples)
	zcr := r.extractor.ExtractZCR(samples, sampleRate)
	centroid, _ := r.extractor.ExtractSpectralCentroid(samples, sampleRate)
	bpm, _ := r.extractor.ExtractBPM(samples, sampleRate)
	chroma, _ := r.extractor.ExtractChroma(samples, sampleRate)
	mfccs, _ := r.extractor.ExtractMFCCs(samples)
	key := r.extractor.ExtractKey(chroma)
	danceability := r.extractor.ExtractDanceability(bpm, energy, zcr)
	acousticness := r.extractor.ExtractAcousticness(centroid, energy)

	return model.TrackFeatures{
		BPM:              bpm,
		Key:              key,
		Energy:           energy,
		Danceability:     danceability,
		Acousticness:     acousticness,
		SpectralCentroid: centroid,
		Chroma:           chroma,
		MFCCs:            mfccs,
		ZCR:              zcr,
	}
}
