package media

import (
	"log/slog"
	"os"

	"github.com/dhowden/tag"
)

//go:generate moq --out ./mocks/metadatareader_mock.go --pkg mocks --skip-ensure --with-resets -fmt goimports . MetadataReader

type MetadataReader interface {
	Metadata(path string) (tag.Metadata, error)
}

type tagReader struct{}

func NewMetadataReader() MetadataReader {
	return &tagReader{}
}

func (*tagReader) Metadata(path string) (tag.Metadata, error) {
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

	md, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	rawdata := md.Raw()
	slog.Debug("metadata", slog.Any("rawdata", rawdata))

	return md, err
	//return tag.ReadFrom(f)
}
