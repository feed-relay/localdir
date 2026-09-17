package fs

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/feed-relay/localdir/internal/media"
)

type Finder struct {
	metadataReader media.MetadataReader
	durationReader media.DurationReader
}

func NewFinder(metadataReader media.MetadataReader, durationReader media.DurationReader) *Finder {
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

	candidates := make([]BaseAudioFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		atype, ok := AudioTypeByName(entry.Name())
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

		candidates = append(candidates, BaseAudioFile{
			Name:      entry.Name(),
			Path:      filepath.Join(absDir, entry.Name()),
			ModTime:   info.ModTime(),
			Length:    info.Size(),
			AudioType: atype,
		})
	}

	slices.SortStableFunc(candidates, func(a, b BaseAudioFile) int {
		if c := b.ModTime.Compare(a.ModTime); c != 0 { // newest first
			return c
		}

		return cmp.Compare(a.Path, b.Path)
	})

	files := make([]AudioFile, 0, min(limit, len(candidates)))
	for _, c := range candidates {
		if len(files) == limit {
			break
		}

		// Probing happens only for the files that can still make the cut,
		// instead of for every file in the directory.
		metadata, err := f.metadataReader.Metadata(c.Path)
		if err != nil {
			slog.Error("fs: read metadata", slog.String("path", c.Path), slog.Any("err", err))
			continue
		}

		duration, err := f.durationReader.Duration(context.Background(), c.Path)
		if err != nil {
			slog.Error("fs: read duration", slog.String("path", c.Path), slog.Any("err", err))
			continue
		}

		files = append(files, AudioFile{
			BaseAudioFile: c,
			Metadata:      metadata,
			Duration:      duration,
		})
	}

	return files, nil
}
