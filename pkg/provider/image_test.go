package provider

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/dhowden/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImage(t *testing.T) {
	t.Run("returns existing PNG without reading picture", func(t *testing.T) {
		dir := t.TempDir()

		const name = "cover"
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, name+".png"),
			[]byte("existing png"),
			0o600,
		))

		got, err := picture(dir, name, nil)

		require.NoError(t, err)
		assert.Equal(t, "cover.png", got)
	})

	t.Run("returns existing JPG when PNG does not exist", func(t *testing.T) {
		dir := t.TempDir()

		const name = "cover"
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, name+".jpg"),
			[]byte("existing jpg"),
			0o600,
		))

		got, err := picture(dir, name, nil)

		require.NoError(t, err)
		assert.Equal(t, "cover.jpg", got)
	})

	t.Run("prefers existing PNG over existing JPG", func(t *testing.T) {
		dir := t.TempDir()

		const name = "cover"

		require.NoError(t, os.WriteFile(
			filepath.Join(dir, name+".png"),
			[]byte("png"),
			0o600,
		))
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, name+".jpg"),
			[]byte("jpg"),
			0o600,
		))

		got, err := picture(dir, name, nil)

		require.NoError(t, err)
		assert.Equal(t, "cover.png", got)
	})

	t.Run("saves PNG", func(t *testing.T) {
		dir := t.TempDir()

		data := []byte("png data")
		pic := &tag.Picture{
			MIMEType: "image/png",
			Data:     data,
		}

		got, err := picture(dir, "cover", pic)

		require.NoError(t, err)
		assert.Equal(t, "cover.png", got)

		saved, err := os.ReadFile(filepath.Join(dir, "cover.png"))
		require.NoError(t, err)
		assert.Equal(t, data, saved)

		info, err := os.Stat(filepath.Join(dir, "cover.png"))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("saves JPEG", func(t *testing.T) {
		dir := t.TempDir()

		data := []byte("jpeg data")
		pic := &tag.Picture{
			MIMEType: "image/jpeg",
			Data:     data,
		}

		got, err := picture(dir, "cover", pic)

		require.NoError(t, err)
		assert.Equal(t, "cover.jpg", got)

		saved, err := os.ReadFile(filepath.Join(dir, "cover.jpg"))
		require.NoError(t, err)
		assert.Equal(t, data, saved)
	})

	t.Run("converts GIF to JPEG", func(t *testing.T) {
		dir := t.TempDir()

		gifData := mustGIF(t)

		pic := &tag.Picture{
			MIMEType: "image/gif",
			Data:     gifData,
		}

		got, err := picture(dir, "cover", pic)

		require.NoError(t, err)
		assert.Equal(t, "cover.jpg", got)

		data, err := os.ReadFile(filepath.Join(dir, "cover.jpg"))
		require.NoError(t, err)

		_, err = jpeg.Decode(bytes.NewReader(data))
		assert.NoError(t, err)
	})

	t.Run("returns error for unsupported image type", func(t *testing.T) {
		dir := t.TempDir()

		pic := &tag.Picture{
			MIMEType: "image/webp",
			Data:     []byte("webp"),
		}

		got, err := picture(dir, "cover", pic)

		require.Error(t, err)
		assert.Empty(t, got)
		assert.ErrorContains(t, err, `unsupported image MIME type: "image/webp"`)
	})

	t.Run("returns error when directory is invalid", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "not-a-directory")

		require.NoError(t, os.WriteFile(path, []byte("file"), 0o600))

		pic := &tag.Picture{
			MIMEType: "image/png",
			Data:     []byte("png"),
		}

		got, err := picture(path, "cover", pic)

		require.Error(t, err)
		assert.Empty(t, got)
	})
}

func TestExistingImage(t *testing.T) {
	t.Run("returns PNG", func(t *testing.T) {
		dir := t.TempDir()

		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "cover.png"),
			[]byte("png"),
			0o600,
		))

		got, ok, err := existingImage(dir, "cover")

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, "cover.png", got)
	})

	t.Run("returns JPG", func(t *testing.T) {
		dir := t.TempDir()

		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "cover.jpg"),
			[]byte("jpg"),
			0o600,
		))

		got, ok, err := existingImage(dir, "cover")

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, "cover.jpg", got)
	})

	t.Run("prefers PNG over JPG", func(t *testing.T) {
		dir := t.TempDir()

		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "cover.png"),
			[]byte("png"),
			0o600,
		))
		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "cover.jpg"),
			[]byte("jpg"),
			0o600,
		))

		got, ok, err := existingImage(dir, "cover")

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, "cover.png", got)
	})

	t.Run("returns false when image does not exist", func(t *testing.T) {
		dir := t.TempDir()

		got, ok, err := existingImage(dir, "cover")

		require.NoError(t, err)
		assert.False(t, ok)
		assert.Empty(t, got)
	})

	t.Run("ignores non-regular files", func(t *testing.T) {
		dir := t.TempDir()

		require.NoError(t, os.Mkdir(
			filepath.Join(dir, "cover.png"),
			0o755,
		))

		require.NoError(t, os.WriteFile(
			filepath.Join(dir, "cover.jpg"),
			[]byte("jpg"),
			0o600,
		))

		got, ok, err := existingImage(dir, "cover")

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, "cover.jpg", got)
	})

	t.Run("returns stat error", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "file")

		require.NoError(t, os.WriteFile(file, []byte("data"), 0o600))

		got, ok, err := existingImage(file, "cover")

		require.Error(t, err)
		assert.False(t, ok)
		assert.Empty(t, got)
	})
}

func TestImageData(t *testing.T) {
	t.Run("returns PNG unchanged", func(t *testing.T) {
		data := []byte("png")

		got, ext, err := imageData(&tag.Picture{
			MIMEType: "image/png",
			Data:     data,
		})

		require.NoError(t, err)
		assert.Equal(t, "png", ext)
		assert.Equal(t, data, got)
	})

	t.Run("returns JPEG unchanged", func(t *testing.T) {
		data := []byte("jpeg")

		got, ext, err := imageData(&tag.Picture{
			MIMEType: "image/jpeg",
			Data:     data,
		})

		require.NoError(t, err)
		assert.Equal(t, "jpg", ext)
		assert.Equal(t, data, got)
	})

	t.Run("returns JPG unchanged", func(t *testing.T) {
		data := []byte("jpg")

		got, ext, err := imageData(&tag.Picture{
			MIMEType: "image/jpg",
			Data:     data,
		})

		require.NoError(t, err)
		assert.Equal(t, "jpg", ext)
		assert.Equal(t, data, got)
	})

	t.Run("converts GIF to JPEG", func(t *testing.T) {
		data := mustGIF(t)

		got, ext, err := imageData(&tag.Picture{
			MIMEType: "image/gif",
			Data:     data,
		})

		require.NoError(t, err)
		assert.Equal(t, "jpg", ext)

		_, err = jpeg.Decode(bytes.NewReader(got))
		assert.NoError(t, err)
	})

	t.Run("returns error for malformed GIF", func(t *testing.T) {
		got, ext, err := imageData(&tag.Picture{
			MIMEType: "image/gif",
			Data:     []byte("not a gif"),
		})

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Empty(t, ext)
	})

	t.Run("returns error for unsupported MIME type", func(t *testing.T) {
		got, ext, err := imageData(&tag.Picture{
			MIMEType: "image/webp",
			Data:     []byte("webp"),
		})

		require.Error(t, err)
		assert.Nil(t, got)
		assert.Empty(t, ext)
		assert.ErrorContains(t, err, `unsupported image MIME type: "image/webp"`)
	})
}

func TestGifToJPEG(t *testing.T) {
	t.Run("converts valid GIF", func(t *testing.T) {
		data := mustGIF(t)

		got, err := gifToJPEG(data)

		require.NoError(t, err)
		require.NotEmpty(t, got)

		_, err = jpeg.Decode(bytes.NewReader(got))
		assert.NoError(t, err)
	})

	t.Run("returns error for invalid GIF", func(t *testing.T) {
		got, err := gifToJPEG([]byte("invalid gif"))

		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func mustGIF(t *testing.T) []byte {
	t.Helper()

	img := image.NewPaletted(
		image.Rect(0, 0, 2, 2),
		color.Palette{
			color.Black,
			color.White,
		},
	)

	img.SetColorIndex(0, 0, 1)
	img.SetColorIndex(1, 0, 1)
	img.SetColorIndex(0, 1, 1)
	img.SetColorIndex(1, 1, 1)

	var buf bytes.Buffer

	require.NoError(t, gif.Encode(&buf, img, nil))

	return buf.Bytes()
}
