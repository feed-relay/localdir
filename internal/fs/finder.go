package fs

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
)

//go:generate moq --out ./mocks/metadatareader_mock.go --pkg mocks --skip-ensure --with-resets -fmt goimports . MetadataReader
//go:generate moq --out ./mocks/durationreader_mock.go --pkg mocks --skip-ensure --with-resets -fmt goimports . DurationReader

type Finder struct {
	metadataReader MetadataReader
	durationReader DurationReader
}

func NewFinder(metadataReader MetadataReader, durationReader DurationReader) *Finder {
	return &Finder{
		metadataReader: metadataReader,
		durationReader: durationReader,
	}
}

// RecentAudioFiles returns the most recently modified audio files in dir,
// sorted newest first, at most limit entries. Files with equal modification
// times are ordered by path, so the result is deterministic.
//
// Audio files are those with extensions .mp3 and .m4a. Subdirectories are not
// traversed. Files whose metadata or duration cannot be read are skipped and
// logged, they do not fail the call.
//
// If limit <= 0 or dir is empty, an empty slice and a nil error are returned.
func (f *Finder) RecentAudioFiles(dir string, limit int) ([]AudioFile, error) {
	if limit <= 0 || dir == "" {
		return []AudioFile{}, nil
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", dir, err)
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, fmt.Errorf("read dir %q: %w", absDir, err)
	}

	candidates := make([]AudioFileBase, 0, len(entries))
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
			// The file may have been removed between ReadDir and Info.
			if !errors.Is(err, os.ErrNotExist) {
				slog.Error("fs: stat", slog.String("name", entry.Name()), slog.Any("err", err))
			}
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}

		candidates = append(candidates, AudioFileBase{
			path:      filepath.Join(absDir, entry.Name()),
			modTime:   info.ModTime(),
			length:    info.Size(),
			audioType: atype,
		})
	}

	slices.SortStableFunc(candidates, func(a, b AudioFileBase) int {
		if c := b.modTime.Compare(a.modTime); c != 0 { // newest first
			return c
		}

		return cmp.Compare(a.path, b.path)
	})

	files := make([]AudioFile, 0, min(limit, len(candidates)))
	for _, c := range candidates {
		if len(files) == limit {
			break
		}

		// Probing happens only for the files that can still make the cut,
		// instead of for every file in the directory.
		metadata, err := f.metadataReader.Read(c.path)
		if err != nil {
			slog.Error("fs: read metadata", slog.String("path", c.path), slog.Any("err", err))
			continue
		}

		duration, err := f.durationReader.Read(c.path)
		if err != nil {
			slog.Error("fs: read duration", slog.String("path", c.path), slog.Any("err", err))
			continue
		}

		files = append(files, AudioFile{
			AudioFileBase: AudioFileBase{
				path:      c.path,
				modTime:   c.modTime,
				length:    c.length,
				audioType: c.audioType,
			},
			metadata: metadata,
			duration: duration,
		})
	}

	return files, nil
}
