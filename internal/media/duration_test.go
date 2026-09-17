package media

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/feed-relay/localdir/internal/media/mocks"
)

func TestDurationReader_Duration(t *testing.T) {
	t.Run("returns error when file does not exist", func(t *testing.T) {
		runner := &mocks.CommandRunnerMock{}
		reader := NewFfprobeDurationReader(runner)

		_, err := reader.Duration(t.Context(), filepath.Join(t.TempDir(), "missing.mp3"))

		require.Error(t, err)
	})

	t.Run("returns error for invalid MP3", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "invalid.mp3")

		err := os.WriteFile(path, []byte("not an mp3 file"), 0o600)
		require.NoError(t, err)

		runner := &mocks.CommandRunnerMock{}
		reader := NewFfprobeDurationReader(runner)

		_, err = reader.Duration(t.Context(), path)

		require.Error(t, err)
	})
}

func TestFFProbeDurationReader_Duration(t *testing.T) {
	t.Parallel()

	const path = "/tmp/test.mp3"

	tests := []struct {
		name         string
		output       []byte
		runnerErr    error
		wantDuration time.Duration
		wantErr      string
	}{
		{
			name:         "success",
			output:       []byte("3.030204\n"),
			wantDuration: 3*time.Second + 30*time.Millisecond + 204*time.Microsecond,
		},
		{
			name:    "invalid duration",
			output:  []byte("not-a-duration\n"),
			wantErr: "parse ffprobe duration",
		},
		{
			name:      "ffprobe error",
			runnerErr: errors.New("exit status 1"),
			wantErr:   `ffprobe "/tmp/test.mp3"`,
		},
		{
			name:    "unknown duration",
			output:  []byte("N/A\n"),
			wantErr: "parse ffprobe duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				gotContext context.Context
				gotName    string
				gotArgs    []string
			)

			runner := &mocks.CommandRunnerMock{RunFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				gotContext = ctx
				gotName = name
				gotArgs = args

				return tt.output, tt.runnerErr
			}}

			reader := NewFfprobeDurationReader(runner)
			ctx := context.WithValue(context.Background(), struct{}{}, "test")

			got, err := reader.Duration(ctx, path)

			require.Equal(t, ctx, gotContext)
			require.Equal(t, "ffprobe", gotName)
			require.Equal(t, []string{
				"-v", "error",
				"-show_entries", "format=duration",
				"-of", "default=noprint_wrappers=1:nokey=1",
				path,
			}, gotArgs)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Zero(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantDuration, got)
		})
	}
}

func TestFFProbeDurationReader_Duration_Testdata(t *testing.T) {
	t.Parallel()

	//if _, err := exec.LookPath("ffprobe"); err != nil {
	//	t.Skip("ffprobe is not installed")
	//}

	reader := NewFfprobeDurationReader(&ExecCommandRunner{})

	got, err := reader.Duration(
		context.Background(),
		"testdata/three_sec_tagged_with_cover.mp3",
	)

	require.NoError(t, err)
	assert.InDelta(t, 3*time.Second, got, float64(100*time.Millisecond))
	//TODO assert.Equal(t, 3030204*time.Microsecond, got)
	//TODO assert.Equal(t, int64(3), int64(got.Seconds()))
}
