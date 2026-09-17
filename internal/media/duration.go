package media

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//go:generate moq --out ./mocks/durationreader_mock.go --pkg mocks --skip-ensure --with-resets -fmt goimports . DurationReader
//go:generate moq --out ./mocks/commandrunner_mock.go --pkg mocks --skip-ensure --with-resets -fmt goimports . CommandRunner

type DurationReader interface {
	Duration(ctx context.Context, path string) (time.Duration, error)
}

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ffprobeDurationReader struct {
	runner CommandRunner
}

type ExecCommandRunner struct{}

func NewFfprobeDurationReader(runner CommandRunner) DurationReader {
	return &ffprobeDurationReader{
		runner: runner,
	}
}

func (r *ffprobeDurationReader) Duration(ctx context.Context, path string) (time.Duration, error) {
	output, err := r.runner.Run(
		ctx,
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	if err != nil {
		return 0, fmt.Errorf("ffprobe %q: %w", path, err)
	}

	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("parse ffprobe duration %q: %w", path, err)
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

func (*ExecCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}
