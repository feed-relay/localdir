package fs

import (
	"log/slog"
	"os"

	"github.com/dhowden/tag"
)

type MetadataReader interface {
	Read(path string) (Metadata, error)
}

type Metadata interface {
	Title() string
	Artist() string
	AlbumArtist() string
	Comment() string
	//Picture() *tag.Picture
}

type tagReader struct{}

func (*tagReader) Read(path string) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			slog.Error("close file", slog.Any("err", err))
		}
	}(f)

	return tag.ReadFrom(f)
}
