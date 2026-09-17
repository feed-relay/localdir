package config

import (
	"log/slog"

	"github.com/meesooqa/go-lgr"

	"github.com/feed-relay/localdir/internal/fs"
)

type AppConfig struct {
	Logger lgr.Config `yaml:"logger"`
	Feeds  []Feed     `yaml:"feeds"`
}

type Feed struct {
	RawSlug        string `yaml:"slug"`
	RawTitle       string `yaml:"title"`
	RawDescription string `yaml:"description"`
	RawLink        string `yaml:"link"`
	RawImage       string `yaml:"image"`
	RawLimit       int    `yaml:"limit"`
	RawDir         string `yaml:"dir"`

	shows      []string
	finder     *fs.Finder
	audioFiles []fs.AudioFile
}

func (f *Feed) Limit() int {
	return f.RawLimit
}

func (f *Feed) Link() string {
	return f.RawLink
}

func (f *Feed) Slug() string {
	return f.RawSlug
}

func (f *Feed) Title() string {
	return f.RawTitle
}

func (f *Feed) Description() string {
	return f.RawDescription
}

func (f *Feed) Image() string {
	return f.RawImage
}

func (f *Feed) Shows() []string {
	if f.shows == nil {
		err := f.loadAudioFiles()
		if err != nil {
			slog.Error("loadAudioFiles", slog.Any("error", err))
			return nil
		}

		shows := make([]string, len(f.audioFiles))
		for i, audioFile := range f.audioFiles {
			shows[i] = audioFile.Path
		}
		f.shows = shows
	}
	return f.shows
}

func (f *Feed) Dir() string {
	return f.RawDir
}

func (f *Feed) AudioFiles() []fs.AudioFile {
	return f.audioFiles
}

func (f *Feed) WithFileProvider(fileProvider *fs.Finder) *Feed {
	f.finder = fileProvider
	return f
}

func (f *Feed) loadAudioFiles() error {
	if f.audioFiles == nil && f.finder != nil {
		files, err := f.finder.RecentAudioFiles(f.Dir(), f.Limit())
		if err != nil {
			return err
		}
		f.audioFiles = files
	}
	return nil
}
