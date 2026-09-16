package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecentFiles(t *testing.T) {
	dir := t.TempDir()

	files := []struct {
		name string
		when time.Time
	}{
		{"old.mp3", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"middle.m4a", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"new.mp3", time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
	}

	for _, file := range files {
		path := filepath.Join(dir, file.name)

		require.NoError(t, os.WriteFile(path, nil, 0o600))
		require.NoError(t, os.Chtimes(path, file.when, file.when))
	}

	require.NoError(t, os.Mkdir(filepath.Join(dir, "subdir"), 0o700))

	p := NewProvider()
	result, err := p.RecentAudioFiles(dir, 2)

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, filepath.Join(dir, "new.mp3"), result[0].path)
	require.Equal(t, Mp3, result[0].audioType)
	require.Equal(t, filepath.Join(dir, "middle.m4a"), result[1].path)
	require.Equal(t, M4a, result[1].audioType)
}

func TestRecentFiles_LimitGreaterThanFileCount(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.mp3"), nil, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.mp3"), nil, 0o600))

	p := NewProvider()
	result, err := p.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	require.Len(t, result, 2)
}

func TestRecentFiles_OnlyAudio(t *testing.T) {
	dir := t.TempDir()

	mp3 := filepath.Join(dir, "audio.mp3")
	m4a := filepath.Join(dir, "audio.m4a")
	jpg := filepath.Join(dir, "cover.jpg")
	txt := filepath.Join(dir, "text.txt")

	for _, path := range []string{mp3, m4a, jpg, txt} {
		require.NoError(t, os.WriteFile(path, nil, 0o600))
	}

	p := NewProvider()
	result, err := p.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	require.Equal(t, 2, len(result))
	require.ElementsMatch(t, []string{mp3, m4a}, []string{result[0].Path(), result[1].Path()})
}

func TestRecentFiles_MP3ExtensionIsCaseInsensitive(t *testing.T) {
	dir := t.TempDir()

	mp3 := filepath.Join(dir, "audio.MP3")

	require.NoError(t, os.WriteFile(mp3, nil, 0o600))

	p := NewProvider()
	result, err := p.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	require.Equal(t, 1, len(result))
	require.Equal(t, mp3, result[0].Path())
}

func TestRecentFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	p := NewProvider()
	result, err := p.RecentAudioFiles(dir, 10)

	require.NoError(t, err)
	require.Empty(t, result)
}

func TestRecentFiles_DirDoesNotExist(t *testing.T) {
	p := NewProvider()
	result, err := p.RecentAudioFiles(filepath.Join(t.TempDir(), "missing"), 10)

	require.Error(t, err)
	require.Nil(t, result)
}

func TestRecentFiles_ZeroLimit(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "file.mp3"),
		nil,
		0o600,
	))

	p := NewProvider()
	result, err := p.RecentAudioFiles(dir, 0)

	require.NoError(t, err)
	require.Empty(t, result)
}
