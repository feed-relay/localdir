package media

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataReader_Metadata(t *testing.T) {
	t.Parallel()

	reader := NewMetadataReader()

	metadata, err := reader.Metadata(
		filepath.Join("testdata", "three_sec_tagged_with_cover.mp3"),
	)

	require.NoError(t, err)
	require.NotNil(t, metadata)

	assert.Equal(t, "Test Audio Track", metadata.Title())
	assert.Equal(t, "Test Artist", metadata.Artist())
	assert.Equal(t, "Test Album Artist", metadata.AlbumArtist())

	picture := metadata.Picture()

	require.NotNil(t, picture)
	assert.Equal(t, "image/jpeg", picture.MIMEType)
	assert.NotEmpty(t, picture.Data)
}

func TestMetadataReader_Metadata_FileNotFound(t *testing.T) {
	t.Parallel()

	reader := NewMetadataReader()

	metadata, err := reader.Metadata(
		filepath.Join("testdata", "does-not-exist.mp3"),
	)

	require.Error(t, err)
	assert.Nil(t, metadata)
}
