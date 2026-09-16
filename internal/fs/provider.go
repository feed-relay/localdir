package fs

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type AudioType string

const (
	M4a AudioType = "m4a"
	Mp3 AudioType = "mp3"
)

type Provider struct {
}

func NewProvider() *Provider {
	return &Provider{}
}

type AudioFile struct {
	path      string
	modTime   time.Time
	length    int64
	audioType AudioType
}

// RecentAudioFiles returns paths to the most recently modified
// audio files in dir, sorted newest first, at most limit entries.

// RecentAudioFiles returns paths to the most recently modified audio
// files in dir, sorted newest first, at most limit entries.
//
// Audio files are those with extensions .mp3, .m4a.
// Subdirectories are not traversed.
//
// If limit <= 0, an empty slice and a nil error are returned.
func (p *Provider) RecentAudioFiles(dir string, limit int) ([]AudioFile, error) {
	if limit <= 0 || dir == "" {
		return []AudioFile{}, nil
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	files := make([]AudioFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		atype, ok := audioType(entry.Name())
		if !ok {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, AudioFile{
			path:      filepath.Join(absDir, entry.Name()),
			modTime:   info.ModTime(),
			length:    info.Size(),
			audioType: atype,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})

	if limit > len(files) {
		limit = len(files)
	}
	return files[:limit], nil
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
