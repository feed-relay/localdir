package fs

import (
	"path/filepath"
	"strings"
	"time"
)

type AudioType string

const (
	M4a AudioType = "m4a"
	Mp3 AudioType = "mp3"
)

type AudioFile struct {
	path      string
	modTime   time.Time
	length    int64
	audioType AudioType
	metadata  Metadata
	duration  int64
}

func (f *AudioFile) Path() string {
	return f.path
}

func (f *AudioFile) ModTime() time.Time {
	return f.modTime
}

func (f *AudioFile) Length() int64 {
	return f.length
}

func (f *AudioFile) AudioType() AudioType {
	return f.audioType
}

func (f *AudioFile) Metadata() Metadata {
	return f.metadata
}

func (f *AudioFile) Duration() int64 {
	return f.duration
}

func audioType(name string) (AudioType, bool) {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mp3":
		return Mp3, true
	case ".m4a":
		return M4a, true
	default:
		return "", false
	}
}
