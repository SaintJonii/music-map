package metadata

import (
	"context"
	"errors"
	"net/http"

	"github.com/SaintJonii/music-map/model"
	"go.uploadedlobster.com/mbtypes"
	"go.uploadedlobster.com/musicbrainzws2"
)

// ErrNotFound is returned when a MusicBrainz entity cannot be found.
var ErrNotFound = errors.New("metadata: entity not found")

// MusicBrainzClient enriches Track metadata via the MusicBrainz API.
type MusicBrainzClient interface {
	LookupByISRC(ctx context.Context, isrc string) (model.Track, error)
	LookupByMBID(ctx context.Context, mbid string) (model.Track, error)
}

// musicBrainzClient wraps the musicbrainzws2 client.
type musicBrainzClient struct {
	client *musicbrainzws2.Client
}

// NewMusicBrainzClient creates a MusicBrainzClient pointing at the given
// MusicBrainz API base URL.  For testing, pass an httptest server URL.
func NewMusicBrainzClient(baseURL string) MusicBrainzClient {
	appInfo := musicbrainzws2.AppInfo{
		Name:    "mapa-musical",
		Version: "0.1.0",
		URL:     "https://github.com/SaintJonii/music-map",
	}
	c := musicbrainzws2.NewClient(appInfo)
	c.SetBaseURL(baseURL)
	return &musicBrainzClient{client: c}
}

func (c *musicBrainzClient) LookupByISRC(ctx context.Context, isrc string) (model.Track, error) {
	result, err := c.client.LookupISRC(ctx, mbtypes.ISRC(isrc), musicbrainzws2.IncludesFilter{})
	if err != nil {
		return model.Track{}, wrapMBError(err)
	}

	if len(result.Recordings) == 0 {
		return model.Track{}, ErrNotFound
	}

	return recordingToTrack(result.Recordings[0]), nil
}

func (c *musicBrainzClient) LookupByMBID(ctx context.Context, mbid string) (model.Track, error) {
	recording, err := c.client.LookupRecording(ctx, mbtypes.MBID(mbid), musicbrainzws2.IncludesFilter{})
	if err != nil {
		return model.Track{}, wrapMBError(err)
	}

	return recordingToTrack(recording), nil
}

// recordingToTrack converts a MusicBrainz Recording to a model.Track.
func recordingToTrack(r musicbrainzws2.Recording) model.Track {
	track := model.Track{
		ID:     string(r.ID),
		Title:  r.Title,
		Artist: r.ArtistCredit.String(),
	}

	// Extract the first release date year if available.
	if r.FirstReleaseDate.Year > 0 {
		track.Year = r.FirstReleaseDate.Year
	}

	// Populate ISRCs.
	if len(r.ISRCs) > 0 {
		track.ISRC = r.ISRCs[0].Compact()
	}

	return track
}

// wrapMBError converts a musicbrainzws2 client error into a sentinel or descriptive error.
func wrapMBError(err error) error {
	var clientErr *musicbrainzws2.ClientError
	if errors.As(err, &clientErr) {
		switch clientErr.StatusCode {
		case http.StatusNotFound:
			return ErrNotFound
		case http.StatusServiceUnavailable:
			return errors.New("metadata: MusicBrainz service unavailable (503)")
		default:
			return errors.New("metadata: MusicBrainz HTTP error: " + clientErr.Error())
		}
	}
	return err
}
