package fs

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

type AudioType string

const (
	M4a AudioType = "m4a"
	Mp3 AudioType = "mp3"
)

type BaseAudioFile struct {
	Name      string
	Path      string
	ModTime   time.Time
	Length    int64
	AudioType AudioType
}

type AudioFile struct {
	BaseAudioFile

	Metadata tag.Metadata
	Duration time.Duration
}

func AudioTypeByName(name string) (AudioType, bool) {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mp3":
		return Mp3, true
	case ".m4a":
		return M4a, true
	default:
		return "", false
	}
}
