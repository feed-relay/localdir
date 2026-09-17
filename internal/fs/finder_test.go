package fs

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/feed-relay/localdir/internal/media"
	mediaMocks "github.com/feed-relay/localdir/internal/media/mocks"
)

var errReaderFailed = errors.New("reader failed")

func TestFinder_RecentAudioFiles_ReturnsNewestFilesFirst(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().Truncate(time.Second)

	writeFile(t, dir, "oldest.mp3", now.Add(-3*time.Hour), 10)
	writeFile(t, dir, "middle.m4a", now.Add(-2*time.Hour), 10)
	writeFile(t, dir, "newest.mp3", now.Add(-1*time.Hour), 10)

	finder, metadataReader, durationReader := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	assert.Equal(t, []string{"newest.mp3", "middle.m4a", "oldest.mp3"}, baseNames(files))
	assert.Len(t, metadataReader.ReadCalls(), 3)
	assert.Len(t, durationReader.ReadCalls(), 3)
}

func TestFinder_RecentAudioFiles_AppliesLimit(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().Truncate(time.Second)

	writeFile(t, dir, "a.mp3", now.Add(-3*time.Hour), 10)
	writeFile(t, dir, "b.mp3", now.Add(-2*time.Hour), 10)
	writeFile(t, dir, "c.mp3", now.Add(-1*time.Hour), 10)

	finder, _, _ := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(dir, 2)

	require.NoError(t, err)
	assert.Equal(t, []string{"c.mp3", "b.mp3"}, baseNames(files))
}

func TestFinder_RecentAudioFiles_LimitGreaterThanNumberOfFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "only.mp3", time.Now(), 10)

	finder, _, _ := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(dir, 100)

	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestFinder_RecentAudioFiles_PopulatesEveryField(t *testing.T) {
	dir := t.TempDir()
	modTime := time.Now().Add(-time.Hour).Truncate(time.Second)
	const size = 1234

	writeFile(t, dir, "episode.m4a", modTime, size)

	finder, _, _ := newFinder(t, map[string]int64{"episode.m4a": 4242})

	files, err := finder.RecentAudioFiles(dir, 1)

	require.NoError(t, err)
	require.Len(t, files, 1)

	file := files[0]
	assert.True(t, filepath.IsAbs(file.Path), "path must be absolute")
	assert.Equal(t, filepath.Join(dir, "episode.m4a"), file.Path)
	assert.Equal(t, int64(size), file.Length)
	assert.Equal(t, M4a, file.AudioType)
	assert.Equal(t, int64(4242), file.Duration)
	assert.WithinDuration(t, modTime, file.ModTime, time.Second)
	// If Metadata carries fields (title, artist, ...), return a distinct value
	// from the mock above and assert it here instead of the zero value.
	assert.Equal(t, &mediaMocks.AudioMetadataMock{}, file.Metadata)
}

func TestFinder_RecentAudioFiles_DetectsBothAudioTypes(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	writeFile(t, dir, "one.mp3", now.Add(-time.Minute), 10)
	writeFile(t, dir, "two.m4a", now.Add(-2*time.Minute), 10)

	finder, _, _ := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	require.Len(t, files, 2)
	assert.Equal(t, Mp3, files[0].AudioType)
	assert.Equal(t, M4a, files[1].AudioType)
}

func TestFinder_RecentAudioFiles_SkipsNonAudioEntries(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	writeFile(t, dir, "notes.txt", now, 10)
	writeFile(t, dir, "cover.jpg", now, 10)
	writeFile(t, dir, "no-extension", now, 10)
	writeFile(t, dir, "episode.mp3", now, 10)

	require.NoError(t, os.Mkdir(filepath.Join(dir, "nested.mp3"), 0o755))
	writeFile(t, filepath.Join(dir, "nested.mp3"), "inner.mp3", now, 10)

	finder, metadataReader, durationReader := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	assert.Equal(t, []string{"episode.mp3"}, baseNames(files))
	assert.Len(t, metadataReader.ReadCalls(), 1, "readers must not be called for skipped entries")
	assert.Len(t, durationReader.ReadCalls(), 1)
}

func TestFinder_RecentAudioFiles_DoesNotTraverseSubdirectories(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	require.NoError(t, os.Mkdir(sub, 0o755))
	writeFile(t, sub, "hidden.mp3", time.Now(), 10)

	finder, _, _ := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestFinder_RecentAudioFiles_SkipsFileWhenMetadataReadFails(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	writeFile(t, dir, "broken.mp3", now.Add(-time.Minute), 10)
	writeFile(t, dir, "good.mp3", now.Add(-2*time.Minute), 10)

	finder, _, durationReader := newFinder(t, nil)
	finder = NewFinder(
		&mediaMocks.MetadataReaderMock{
			ReadFunc: func(path string) (media.AudioMetadata, error) {
				if filepath.Base(path) == "broken.mp3" {
					return &mediaMocks.AudioMetadataMock{}, errReaderFailed
				}
				return &mediaMocks.AudioMetadataMock{}, nil
			},
		},
		durationReader,
	)

	files, err := finder.RecentAudioFiles(dir, 10)

	require.NoError(t, err, "a failing file must not fail the whole call")
	assert.Equal(t, []string{"good.mp3"}, baseNames(files))
	assert.Len(t, durationReader.ReadCalls(), 1, "duration must not be read after metadata failed")
}

func TestFinder_RecentAudioFiles_SkipsFileWhenDurationReadFails(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	writeFile(t, dir, "broken.mp3", now.Add(-time.Minute), 10)
	writeFile(t, dir, "good.mp3", now.Add(-2*time.Minute), 10)

	finder := NewFinder(
		&mediaMocks.MetadataReaderMock{
			ReadFunc: func(path string) (media.AudioMetadata, error) { return &mediaMocks.AudioMetadataMock{}, nil },
		},
		&mediaMocks.DurationReaderMock{
			ReadFunc: func(path string) (int64, error) {
				if filepath.Base(path) == "broken.mp3" {
					return 0, errReaderFailed
				}
				return 1, nil
			},
		},
	)

	files, err := finder.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	assert.Equal(t, []string{"good.mp3"}, baseNames(files))
}

func TestFinder_RecentAudioFiles_NonPositiveLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{name: "zero", limit: 0},
		{name: "negative", limit: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "episode.mp3", time.Now(), 10)

			finder, metadataReader, durationReader := newFinder(t, nil)

			files, err := finder.RecentAudioFiles(dir, tt.limit)

			require.NoError(t, err)
			assert.NotNil(t, files, "documented to return an empty slice, not nil")
			assert.Empty(t, files)
			assert.Empty(t, metadataReader.ReadCalls())
			assert.Empty(t, durationReader.ReadCalls())
		})
	}
}

func TestFinder_RecentAudioFiles_EmptyDir(t *testing.T) {
	finder, metadataReader, _ := newFinder(t, nil)

	files, err := finder.RecentAudioFiles("", 10)

	require.NoError(t, err)
	assert.NotNil(t, files)
	assert.Empty(t, files)
	assert.Empty(t, metadataReader.ReadCalls())
}

func TestFinder_RecentAudioFiles_EmptyDirectory(t *testing.T) {
	finder, _, _ := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(t.TempDir(), 10)

	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestFinder_RecentAudioFiles_MissingDirectory(t *testing.T) {
	finder, _, _ := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(filepath.Join(t.TempDir(), "does-not-exist"), 10)

	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Empty(t, files)
}

func TestFinder_RecentAudioFiles_DirIsAFile(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "episode.mp3", time.Now(), 10)

	finder, _, _ := newFinder(t, nil)

	_, err := finder.RecentAudioFiles(path, 10)

	require.Error(t, err)
}

func TestFinder_RecentAudioFiles_ResolvesRelativeDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "episode.mp3", time.Now(), 10)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	relDir, err := filepath.Rel(cwd, dir)
	require.NoError(t, err)

	finder, _, _ := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(relDir, 10)

	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, filepath.Join(dir, "episode.mp3"), files[0].Path)
}

func TestFinder_RecentAudioFiles_PassesAbsolutePathsToReaders(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "episode.mp3", time.Now(), 10)

	var metadataPaths, durationPaths []string
	finder := NewFinder(
		&mediaMocks.MetadataReaderMock{
			ReadFunc: func(path string) (media.AudioMetadata, error) {
				metadataPaths = append(metadataPaths, path)
				return &mediaMocks.AudioMetadataMock{}, nil
			},
		},
		&mediaMocks.DurationReaderMock{
			ReadFunc: func(path string) (int64, error) {
				durationPaths = append(durationPaths, path)
				return 0, nil
			},
		},
	)

	_, err := finder.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(dir, "episode.mp3")}, metadataPaths)
	assert.Equal(t, []string{filepath.Join(dir, "episode.mp3")}, durationPaths)
}

func TestFinder_RecentAudioFiles_KeepsDurationPerFile(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	writeFile(t, dir, "first.mp3", now.Add(-time.Minute), 10)
	writeFile(t, dir, "second.mp3", now.Add(-2*time.Minute), 10)

	finder, _, _ := newFinder(t, map[string]int64{
		"first.mp3":  111,
		"second.mp3": 222,
	})

	files, err := finder.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	require.Len(t, files, 2)
	assert.Equal(t, int64(111), files[0].Duration)
	assert.Equal(t, int64(222), files[1].Duration)
}

// TestFinder_RecentAudioFiles_UppercaseExtension documents a gap rather than
// current behaviour: extensions are matched case-sensitively today, so an
// ".MP3" file is silently ignored. Drop the Skip once audioType lowercases the
// extension.
func TestFinder_RecentAudioFiles_UppercaseExtension(t *testing.T) {
	t.Skip("enable after making extension matching case-insensitive")

	dir := t.TempDir()
	writeFile(t, dir, "episode.MP3", time.Now(), 10)

	finder, _, _ := newFinder(t, nil)

	files, err := finder.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	assert.Len(t, files, 1)
}

// newFinder builds a Finder with happy-path mocks. durations maps a file base
// name to the duration the DurationReader should return for it.
func newFinder(t *testing.T, durations map[string]int64) (*Finder, *mediaMocks.MetadataReaderMock, *mediaMocks.DurationReaderMock) {
	t.Helper()

	metadataReader := &mediaMocks.MetadataReaderMock{
		ReadFunc: func(path string) (media.AudioMetadata, error) {
			return &mediaMocks.AudioMetadataMock{}, nil
		},
	}
	durationReader := &mediaMocks.DurationReaderMock{
		ReadFunc: func(path string) (int64, error) {
			return durations[filepath.Base(path)], nil
		},
	}

	return NewFinder(metadataReader, durationReader), metadataReader, durationReader
}

// writeFile creates a file of the given size and modification time.
func writeFile(t *testing.T, dir, name string, modTime time.Time, size int) string {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, bytes.Repeat([]byte("a"), size), 0o600))
	require.NoError(t, os.Chtimes(path, modTime, modTime))

	return path
}

func baseNames(files []AudioFile) []string {
	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, filepath.Base(file.Path))
	}

	return names
}
