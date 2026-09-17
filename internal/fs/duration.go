package fs

import (
	"log/slog"
	"os"

	"github.com/hajimehoshi/go-mp3"
)

type DurationReader interface {
	Read(path string) (int64, error)
}

type durReader struct{}

func (*durReader) Read(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			slog.Error("close file", slog.Any("err", err))
		}
	}(f)

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return 0, err
	}

	durationSeconds := float64(d.Length()) / float64(d.SampleRate())
	return int64(durationSeconds), nil
}
