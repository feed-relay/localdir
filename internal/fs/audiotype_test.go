package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAudioType(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     AudioType
		wantOK   bool
	}{
		{name: "mp3", fileName: "episode.mp3", want: Mp3, wantOK: true},
		{name: "m4a", fileName: "episode.m4a", want: M4a, wantOK: true},
		{name: "several dots", fileName: "episode.2024-01-01.mp3", want: Mp3, wantOK: true},
		{name: "unsupported extension", fileName: "episode.wav", wantOK: false},
		{name: "no extension", fileName: "episode", wantOK: false},
		{name: "extension in the middle", fileName: "episode.mp3.txt", wantOK: false},
		{name: "empty name", fileName: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := AudioTypeByName(tt.fileName)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestAudioType_CaseInsensitive documents a gap: extensions are matched
// case-sensitively today. Drop the Skip once audioType lowercases the
// extension before comparing.
func TestAudioType_CaseInsensitive(t *testing.T) {
	t.Skip("enable after making extension matching case-insensitive")

	got, ok := AudioTypeByName("EPISODE.MP3")

	assert.True(t, ok)
	assert.Equal(t, Mp3, got)
}
