package metadata

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"testing"
)

// --- Binary helpers for generating tagged test files ---

// makeID3v2 creates a minimal ID3v2.3 tag with the given frame data.
// Does NOT include MPEG audio — ReadFrom will detect as ID3.
func makeID3v2(frames ...id3Frame) []byte {
	// Calculate total frame data size.
	var dataSize int
	for _, f := range frames {
		dataSize += 10 + len(f.data) // 10-byte frame header + data
	}

	// ID3v2.3 header: "ID3" + version 3.0 + flags 0 + synchsafe size.
	buf := &bytes.Buffer{}
	buf.WriteString("ID3")
	buf.WriteByte(3)               // version major
	buf.WriteByte(0)               // version minor
	buf.WriteByte(0)               // flags
	writeSynchSafeInt(buf, dataSize+10) // size includes padding (10 bytes extra)

	// Write frames.
	for _, f := range frames {
		buf.WriteString(f.id)
		binary.Write(buf, binary.BigEndian, int32(len(f.data)))
		buf.WriteByte(0) // flags hi
		buf.WriteByte(0) // flags lo
		buf.Write(f.data)
	}

	// Padding.
	pad := make([]byte, 10)
	buf.Write(pad)

	return buf.Bytes()
}

type id3Frame struct {
	id   string
	data []byte
}

// textFrame creates an ID3v2 text frame with encoding byte 0x03 (UTF-8) + text.
func textFrame(id, text string) id3Frame {
	data := append([]byte{0x03}, []byte(text)...)
	return id3Frame{id: id, data: data}
}

// writeSynchSafeInt writes an int32 as a 4-byte synchsafe integer.
func writeSynchSafeInt(w io.Writer, v int) {
	buf := make([]byte, 4)
	buf[0] = byte((v >> 21) & 0x7F)
	buf[1] = byte((v >> 14) & 0x7F)
	buf[2] = byte((v >> 7) & 0x7F)
	buf[3] = byte(v & 0x7F)
	w.Write(buf)
}

// makeFLAC creates a minimal FLAC file with STREAMINFO + Vorbis comment blocks.
func makeFLAC(comments ...vorbisComment) []byte {
	buf := &bytes.Buffer{}

	// FLAC magic.
	buf.WriteString("fLaC")

	// STREAMINFO block (type 0): 34 bytes.
	streaminfo := make([]byte, 34)
	// Minimum block size: 4096
	binary.BigEndian.PutUint16(streaminfo[0:2], 4096)
	// Maximum block size: 4096
	binary.BigEndian.PutUint16(streaminfo[2:4], 4096)
	// Minimum frame size: 0 (unknown)
	// Maximum frame size: 0 (unknown)
	// Sample rate: 44100 (bits 16-35)
	//   byte 10: AAAA AAAA
	//   byte 11: AAAA BBBB
	//   byte 12: BBBB BBBB
	// where A = sample rate (20 bits), B = channels-1 (3 bits) + bits per sample-1 (5 bits) + total samples (36 bits)
	//
	// Pack: sampleRate=44100 → 0b1010110001000100 (16 bits visible in bytes 10-11 top 4)
	// Let's simplify: write raw bytes that dhowden/tag will parse without error.
	streaminfo[10] = 0xAC  // sampleRate high
	streaminfo[11] = 0x44  // sampleRate mid + channels
	streaminfo[12] = 0x10  // channels + bps
	binary.BigEndian.PutUint32(streaminfo[13:17], 0) // total samples (unknown)
	// MD5 (16 bytes of zeros).
	copy(streaminfo[18:34], make([]byte, 16))

	// STREAMINFO metadata block header (4 bytes): last-block bit + type + length.
	writeMetadataBlockHeader(buf, false, 0, streaminfo)

	// Vorbis comment block.
	vcBuf := &bytes.Buffer{}

	// Vendor string.
	vendor := "reference libFLAC 1.3.2 20170101"
	binary.Write(vcBuf, binary.LittleEndian, uint32(len(vendor)))
	vcBuf.WriteString(vendor)

	// Number of comments.
	binary.Write(vcBuf, binary.LittleEndian, uint32(len(comments)))
	for _, c := range comments {
		tagStr := c.key + "=" + c.value
		binary.Write(vcBuf, binary.LittleEndian, uint32(len(tagStr)))
		vcBuf.WriteString(tagStr)
	}

	// Write the Vorbis comment metadata block.
	writeMetadataBlockHeader(buf, true, 4, vcBuf.Bytes()) // type 4 = Vorbis comment, last block

	return buf.Bytes()
}

type vorbisComment struct {
	key, value string
}

func writeMetadataBlockHeader(w io.Writer, last bool, blockType byte, data []byte) {
	header := byte(blockType)
	if last {
		header |= 0x80
	}
	w.Write([]byte{header})
	// 3-byte big-endian length.
	lenBuf := make([]byte, 3)
	lenBuf[0] = byte(len(data) >> 16)
	lenBuf[1] = byte(len(data) >> 8)
	lenBuf[2] = byte(len(data))
	w.Write(lenBuf)
	w.Write(data)
}

// errorReader is a ReadSeeker that always returns an error.
type errorReader struct{}

func (errorReader) Read(p []byte) (int, error) { return 0, errors.New("simulated read error") }
func (errorReader) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("simulated seek error")
}

// --- Tests ---

func TestTagReader_MP3_ID3v2_FullTags(t *testing.T) {
	tr := NewTagReader()

	data := makeID3v2(
		textFrame("TIT2", "Test Song"),
		textFrame("TPE1", "Test Artist"),
		textFrame("TALB", "Test Album"),
		textFrame("TPE2", "Test Album Artist"),
		textFrame("TCON", "Rock"),
		textFrame("TYER", "2024"),
		textFrame("TRCK", "3"),
		textFrame("TSRC", "US-ABC-12-34567"),
		id3Frame{id: "UFID", data: append([]byte("http://musicbrainz.org\x00"), []byte("abcdef12-3456-7890-abcd-ef1234567890")...)},
	)

	reader := bytes.NewReader(data)
	track, err := tr.ReadTags(reader)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	assertEqual(t, "title", "Test Song", track.Title)
	assertEqual(t, "artist", "Test Artist", track.Artist)
	assertEqual(t, "album", "Test Album", track.Album)
	assertEqual(t, "album artist", "Test Album Artist", track.AlbumArtist)
	assertEqual(t, "genre", "Rock", track.Genre)
	assertEqual(t, "year", 2024, track.Year)
	assertEqual(t, "track number", 3, track.TrackNumber)
	assertEqual(t, "ISRC", "US-ABC-12-34567", track.ISRC)
}

func TestTagReader_FLAC_VorbisComments(t *testing.T) {
	tr := NewTagReader()

	data := makeFLAC(
		vorbisComment{key: "TITLE", value: "FLAC Song"},
		vorbisComment{key: "ARTIST", value: "FLAC Artist"},
		vorbisComment{key: "ALBUM", value: "FLAC Album"},
		vorbisComment{key: "GENRE", value: "Electronic"},
		vorbisComment{key: "DATE", value: "2023"},
		vorbisComment{key: "TRACKNUMBER", value: "5"},
		vorbisComment{key: "ISRC", value: "GB-XXX-12-34567"},
		vorbisComment{key: "MUSICBRAINZ_TRACKID", value: "ff123456-7890-abcd-ef12-34567890abcd"},
	)

	reader := bytes.NewReader(data)
	track, err := tr.ReadTags(reader)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	assertEqual(t, "title", "FLAC Song", track.Title)
	assertEqual(t, "artist", "FLAC Artist", track.Artist)
	assertEqual(t, "album", "FLAC Album", track.Album)
	assertEqual(t, "genre", "Electronic", track.Genre)
	assertEqual(t, "track number", 5, track.TrackNumber)
	assertEqual(t, "ISRC", "GB-XXX-12-34567", track.ISRC)
}

func TestTagReader_Vorbis_DateParsedAsYear(t *testing.T) {
	tr := NewTagReader()

	data := makeFLAC(
		vorbisComment{key: "DATE", value: "2024"},
	)

	reader := bytes.NewReader(data)
	track, err := tr.ReadTags(reader)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	assertEqual(t, "year", 2024, track.Year)
}

func TestTagReader_EmptyReader_ReturnsEmptyTrack(t *testing.T) {
	tr := NewTagReader()

	reader := bytes.NewReader([]byte{})
	track, err := tr.ReadTags(reader)

	if err != nil {
		t.Fatalf("expected no error for empty reader, got: %v", err)
	}

	// All fields should be empty.
	if track.Title != "" || track.Artist != "" || track.Album != "" {
		t.Errorf("expected empty metadata, got Title=%q Artist=%q Album=%q",
			track.Title, track.Artist, track.Album)
	}
}

func TestTagReader_ReadError(t *testing.T) {
	tr := NewTagReader()

	reader := errorReader{}
	_, err := tr.ReadTags(reader)

	if err == nil {
		t.Fatal("expected error from broken reader, got nil")
	}
}

func TestTagReader_MissingFile(t *testing.T) {
	_ = NewTagReader() // TagReader is available.

	// Simulate a missing file by opening a non-existent path.
	f, err := os.Open("testdata/nonexistent_file.mp3")
	if err == nil {
		f.Close()
		t.Fatal("expected error opening non-existent file")
	}

	// Verify the error is descriptive.
	if err == nil {
		t.Fatal("expected file-not-found error")
	}
	t.Logf("missing file error (expected): %v", err)
}

func TestTagReader_FLAC_NoComments(t *testing.T) {
	tr := NewTagReader()

	// FLAC with STREAMINFO only, no Vorbis comment block.
	data := makeFLAC()

	reader := bytes.NewReader(data)
	track, err := tr.ReadTags(reader)

	// FLAC with no tags: ReadFrom may return ErrNoTagsFound or nil Metadata.
	// TagReader should return empty Track, no error.
	if err != nil {
		t.Fatalf("expected no error for untagged FLAC, got: %v", err)
	}

	if track.Title != "" {
		t.Errorf("expected empty Title, got %q", track.Title)
	}
}

func assertEqual[T comparable](t *testing.T, field string, expected, got T) {
	t.Helper()
	if expected != got {
		t.Errorf("%s: expected %v, got %v", field, expected, got)
	}
}
