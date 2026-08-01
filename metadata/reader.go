package metadata

import (
	"io"
	"strings"

	"github.com/SaintJonii/music-map/model"
	"github.com/dhowden/tag"
)

// TagReader reads audio metadata tags from an io.ReadSeeker.
type TagReader interface {
	ReadTags(r io.ReadSeeker) (model.Track, error)
}

// NewTagReader creates a new TagReader using dhowden/tag.
func NewTagReader() TagReader {
	return &tagReader{}
}

type tagReader struct{}

func (tr *tagReader) ReadTags(r io.ReadSeeker) (model.Track, error) {
	m, err := tag.ReadFrom(r)
	if err != nil {
		// ErrNoTagsFound and io.EOF mean no tags were found — not an error.
		if err == tag.ErrNoTagsFound || err == io.EOF {
			return model.Track{}, nil
		}
		return model.Track{}, err
	}

	trackNum, _ := m.Track()
	track := model.Track{
		Title:       m.Title(),
		Artist:      m.Artist(),
		Album:       m.Album(),
		AlbumArtist: m.AlbumArtist(),
		Genre:       m.Genre(),
		Year:        m.Year(),
		TrackNumber: trackNum,
	}

	// Extract ISRC from raw tags.
	raw := m.Raw()
	track.ISRC = extractTagString(raw, "ISRC", "TSRC", "tsrc")

	return track, nil
}

// extractTagString searches the raw tag map for a string value under any of the given keys.
func extractTagString(raw map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		for rawKey, v := range raw {
			if strings.EqualFold(rawKey, k) {
				switch val := v.(type) {
				case string:
					return val
				case []byte:
					return string(val)
				case []string:
					if len(val) > 0 {
						return val[0]
					}
				}
			}
		}
	}
	return ""
}
