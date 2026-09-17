package provider

import (
	"bytes"
	"errors"
	"fmt"
	"image/gif"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

func picture(dir, name string, pic *tag.Picture) (string, error) {
	if filename, ok, err := existingImage(dir, name); err != nil {
		return "", err
	} else if ok {
		return filename, nil
	}

	data, ext, err := imageData(pic)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s.%s", name, ext)
	path := filepath.Join(dir, filename)

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o600,
	)
	if err != nil {
		return "", err
	}

	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", err
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}

	return filename, nil
}

func existingImage(dir, name string) (string, bool, error) {
	for _, ext := range []string{"png", "jpg"} {
		filename := fmt.Sprintf("%s.%s", name, ext)
		path := filepath.Join(dir, filename)

		info, err := os.Stat(path)
		if err == nil {
			if info.Mode().IsRegular() {
				return filename, true, nil
			}

			continue
		}

		if !errors.Is(err, os.ErrNotExist) {
			return "", false, err
		}
	}

	return "", false, nil
}

func imageData(pic *tag.Picture) ([]byte, string, error) {
	mime := strings.ToLower(pic.MIMEType)

	switch {
	case strings.Contains(mime, "png"):
		return pic.Data, "png", nil

	case strings.Contains(mime, "jpeg"), strings.Contains(mime, "jpg"):
		return pic.Data, "jpg", nil

	case strings.Contains(mime, "gif"):
		data, err := gifToJPEG(pic.Data)
		if err != nil {
			return nil, "", err
		}
		return data, "jpg", nil

	default:
		return nil, "", fmt.Errorf("unsupported image MIME type: %q", pic.MIMEType)
	}
}

func gifToJPEG(data []byte) ([]byte, error) {
	src, err := gif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
