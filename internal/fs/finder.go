package fs

import (
	"log"
	"os"
	"path/filepath"
	"sort"
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

// RecentAudioFiles returns []AudioFile of the most recently modified audio
// files in dir, sorted newest first, at most limit entries.
//
// Audio files are those with extensions .mp3, .m4a.
// Subdirectories are not traversed.
//
// If limit <= 0, an empty slice and a nil error are returned.
func (f *Finder) RecentAudioFiles(dir string, limit int) ([]AudioFile, error) {
	files, err := f.findRecentAudioFiles(dir, limit)
	if err != nil {
		return files, err
	}
	return files, nil
}

func (f *Finder) findRecentAudioFiles(dir string, limit int) ([]AudioFile, error) {
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

		metadata, err := f.metadataReader.Read(filepath.Join(absDir, entry.Name()))
		if err != nil {
			continue
		}
		duration, err := f.durationReader.Read(filepath.Join(absDir, entry.Name()))
		if err != nil {
			continue
		}

		files = append(files, AudioFile{
			path:      filepath.Join(absDir, entry.Name()),
			modTime:   info.ModTime(),
			length:    info.Size(),
			audioType: atype,
			metadata:  metadata,
			duration:  duration,
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
