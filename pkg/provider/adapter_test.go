package provider

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/dhowden/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/feed-relay/rsscast"

	"github.com/feed-relay/localdir/internal/fs"
	"github.com/feed-relay/localdir/pkg/provider/mocks"
)

type metadataStub struct {
	title       string
	artist      string
	albumArtist string
	comment     string
}

func (m metadataStub) Format() tag.Format          { return tag.UnknownFormat }
func (m metadataStub) FileType() tag.FileType      { return tag.UnknownFileType }
func (m metadataStub) Title() string               { return m.title }
func (m metadataStub) Album() string               { return "" }
func (m metadataStub) Artist() string              { return m.artist }
func (m metadataStub) AlbumArtist() string         { return m.albumArtist }
func (m metadataStub) Composer() string            { return "" }
func (m metadataStub) Genre() string               { return "" }
func (m metadataStub) Year() int                   { return 0 }
func (m metadataStub) Track() (int, int)           { return 0, 0 }
func (m metadataStub) Disc() (int, int)            { return 0, 0 }
func (m metadataStub) Picture() *tag.Picture       { return nil }
func (m metadataStub) Lyrics() string              { return "" }
func (m metadataStub) Comment() string             { return m.comment }
func (m metadataStub) Raw() map[string]interface{} { return nil }

func TestAdapter_Feed(t *testing.T) {
	config := &mocks.ConfigMock{
		GeneratorFunc:        func() string { return "test-generator" },
		ItunesOwnerNameFunc:  func() string { return "John Doe" },
		ItunesOwnerEmailFunc: func() string { return "john@example.com" },
	}

	adapter := &adapter{config: config}
	modTime := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	feed := &mocks.FeedMock{
		LinkFunc:        func() string { return "https://example.com/podcast" },
		ImageFunc:       func() string { return "cover.jpg" },
		TitleFunc:       func() string { return "Test Podcast" },
		DescriptionFunc: func() string { return "Podcast description" },
		AudioFilesFunc: func() []fs.AudioFile {
			return []fs.AudioFile{
				{
					BaseAudioFile: fs.BaseAudioFile{
						Name:      "first.mp3",
						AudioType: fs.Mp3,
						Length:    123,
						ModTime:   modTime,
					},
					Duration: 90 * time.Second,
					Metadata: metadataStub{
						title:   "First episode",
						artist:  "Artist",
						comment: "First description",
					},
				},
				{
					BaseAudioFile: fs.BaseAudioFile{
						Name:      "second.m4a",
						AudioType: fs.M4a,
						Length:    456,
						ModTime:   modTime.Add(time.Hour),
					},
					Duration: 120 * time.Second,
					Metadata: metadataStub{
						title:       "",
						artist:      "",
						albumArtist: "Album Artist",
					},
				},
			}
		},
	}

	got, err := adapter.Feed(context.Background(), feed)
	require.NoError(t, err)
	require.NotNil(t, got)

	var buf bytes.Buffer
	require.NoError(t, got.Encode(&buf))
	xml := buf.String()

	assert.Contains(t, xml, "<title>Test Podcast</title>")
	assert.Contains(t, xml, "<description><![CDATA[Podcast description]]></description>")
	assert.Contains(t, xml, "<generator>test-generator</generator>")
	assert.Contains(t, xml, "<itunes:name>John Doe</itunes:name>")
	assert.Contains(t, xml, "<itunes:email>john@example.com</itunes:email>")
	assert.Contains(t, xml, "<itunes:image href=\"https://example.com/podcast/cover.jpg\"")

	assert.Contains(t, xml, "<title>First episode</title>")
	assert.Contains(t, xml, "<guid>first.mp3</guid>")
	assert.Contains(t, xml, "https://example.com/podcast/first.mp3")
	assert.Contains(t, xml, "audio/mpeg")
	assert.Contains(t, xml, "<itunes:author>Artist</itunes:author>")
	assert.Contains(t, xml, "<description><![CDATA[First description]]></description>")

	assert.Contains(t, xml, "<title>second.m4a</title>")
	assert.Contains(t, xml, "<guid>second.m4a</guid>")
	assert.Contains(t, xml, "https://example.com/podcast/second.m4a")
	assert.Contains(t, xml, "audio/x-m4a")
	assert.Contains(t, xml, "<itunes:author>Album Artist</itunes:author>")
}

func TestAdapter_Feed_UsesFallbacks(t *testing.T) {
	config := &mocks.ConfigMock{
		GeneratorFunc:        func() string { return "generator" },
		ItunesOwnerNameFunc:  func() string { return "owner" },
		ItunesOwnerEmailFunc: func() string { return "owner@example.com" },
	}
	adapter := &adapter{config: config}

	feed := &mocks.FeedMock{
		LinkFunc:        func() string { return "https://example.com/feed" },
		ImageFunc:       func() string { return "" },
		TitleFunc:       func() string { return "" },
		DescriptionFunc: func() string { return "" },
		AudioFilesFunc: func() []fs.AudioFile {
			return []fs.AudioFile{
				{
					BaseAudioFile: fs.BaseAudioFile{
						Name:      "episode.mp3",
						AudioType: fs.Mp3,
					},
					Metadata: metadataStub{},
				},
			}
		},
	}

	got, err := adapter.Feed(context.Background(), feed)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, got.Encode(&buf))
	xml := buf.String()

	assert.Contains(t, xml, "<title>Feed Relay Localdir</title>")
	assert.Contains(t, xml, "<description><![CDATA[Feed Relay Localdir]]></description>")
	assert.Contains(t, xml, "<title>episode.mp3</title>")
	assert.Contains(t, xml, "<itunes:author>Feed Relay Localdir</itunes:author>")
	assert.Contains(t, xml, "<description><![CDATA[episode.mp3]]></description>")
}

func TestEnclosureTypeByAudioType(t *testing.T) {
	tests := []struct {
		name      string
		audioType fs.AudioType
		wantType  rsscast.EnclosureType
		wantOK    bool
	}{
		{name: "mp3", audioType: fs.Mp3, wantType: rsscast.Mp3, wantOK: true},
		{name: "m4a", audioType: fs.M4a, wantType: rsscast.M4a, wantOK: true},
		{name: "unknown", audioType: fs.AudioType("wav"), wantType: rsscast.Mp3, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotOK := enclosureTypeByAudioType(tt.audioType)
			assert.Equal(t, tt.wantType, gotType)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}
