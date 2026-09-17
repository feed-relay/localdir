package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/feed-relay/contracts"
	"github.com/feed-relay/rsscast"

	"github.com/feed-relay/localdir/internal/fs"
	"github.com/feed-relay/localdir/pkg/provider/mocks"
)

func TestProvider_Platform(t *testing.T) {
	p := &Provider{}

	assert.Equal(t, Platform, p.Platform())
}

func TestProvider_Feeds(t *testing.T) {
	feed1 := &mocks.FeedMock{
		SlugFunc:       func() string { return "first" },
		ShowsFunc:      func() []string { return []string{"show-1"} },
		SetFinderFunc:  func(*fs.Finder) {},
		AudioFilesFunc: func() []fs.AudioFile { return nil },
	}
	feed2 := &mocks.FeedMock{
		SlugFunc:       func() string { return "second" },
		ShowsFunc:      func() []string { return []string{"show-2"} },
		SetFinderFunc:  func(*fs.Finder) {},
		AudioFilesFunc: func() []fs.AudioFile { return nil },
	}

	adapter := &mocks.AdapterMock{
		FeedFunc: func(_ context.Context, feed contracts.Feed) (*rsscast.Feed, error) {
			return rsscast.NewFeed(rsscast.FeedData{Title: feed.Slug()}), nil
		},
	}

	p := &Provider{
		adapter: adapter,
	}

	got, err := p.Feeds(context.Background(), []contracts.Feed{feed1, feed2})

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Contains(t, got, "localdir-first")
	assert.Contains(t, got, "localdir-second")
	assert.Equal(t, 2, len(adapter.FeedCalls()))
	assert.Len(t, feed1.SetFinderCalls(), 1)
	assert.Len(t, feed2.SetFinderCalls(), 1)
	assert.Len(t, feed1.ShowsCalls(), 1)
	assert.Len(t, feed2.ShowsCalls(), 1)
}

func TestProvider_Feeds_SkipsFeedsWithoutShows(t *testing.T) {
	feed := &mocks.FeedMock{
		ShowsFunc:      func() []string { return nil },
		SetFinderFunc:  func(*fs.Finder) {},
		AudioFilesFunc: func() []fs.AudioFile { return nil },
	}
	adapter := &mocks.AdapterMock{
		FeedFunc: func(context.Context, contracts.Feed) (*rsscast.Feed, error) {
			t.Fatal("adapter must not be called")
			return nil, nil
		},
	}

	p := &Provider{adapter: adapter}

	got, err := p.Feeds(context.Background(), []contracts.Feed{feed})

	assert.Nil(t, got)
	assert.EqualError(t, err, "localdir: no shows")
	assert.Equal(t, 0, len(adapter.FeedCalls()))
}

func TestProvider_Feeds_ReturnsSuccessfulFeedsAndErrors(t *testing.T) {
	errFailed := errors.New("build failed")

	failedFeed := &mocks.FeedMock{
		SlugFunc:       func() string { return "failed" },
		ShowsFunc:      func() []string { return []string{"show"} },
		SetFinderFunc:  func(*fs.Finder) {},
		AudioFilesFunc: func() []fs.AudioFile { return nil },
	}
	successFeed := &mocks.FeedMock{
		SlugFunc:       func() string { return "success" },
		ShowsFunc:      func() []string { return []string{"show"} },
		SetFinderFunc:  func(*fs.Finder) {},
		AudioFilesFunc: func() []fs.AudioFile { return nil },
	}

	successRSS := rsscast.NewFeed(rsscast.FeedData{Title: "success"})
	adapter := &mocks.AdapterMock{
		FeedFunc: func(_ context.Context, feed contracts.Feed) (*rsscast.Feed, error) {
			switch feed {
			case failedFeed:
				return nil, errFailed
			case successFeed:
				return successRSS, nil
			default:
				return nil, errors.New("unexpected feed")
			}
		},
	}

	p := &Provider{adapter: adapter}

	got, err := p.Feeds(context.Background(), []contracts.Feed{failedFeed, successFeed})

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	require.Len(t, got, 1)
	assert.Same(t, successRSS, got["localdir-success"])
}

func TestProvider_Feeds_ReturnsAllErrorsWhenAllFeedsFail(t *testing.T) {
	errFirst := errors.New("first failed")
	errSecond := errors.New("second failed")

	feed1 := &mocks.FeedMock{
		SlugFunc:       func() string { return "first" },
		ShowsFunc:      func() []string { return []string{"show"} },
		SetFinderFunc:  func(*fs.Finder) {},
		AudioFilesFunc: func() []fs.AudioFile { return nil },
	}
	feed2 := &mocks.FeedMock{
		SlugFunc:       func() string { return "second" },
		ShowsFunc:      func() []string { return []string{"show"} },
		SetFinderFunc:  func(*fs.Finder) {},
		AudioFilesFunc: func() []fs.AudioFile { return nil },
	}

	adapter := &mocks.AdapterMock{
		FeedFunc: func(_ context.Context, feed contracts.Feed) (*rsscast.Feed, error) {
			if feed == feed1 {
				return nil, errFirst
			}
			return nil, errSecond
		},
	}

	p := &Provider{adapter: adapter}

	got, err := p.Feeds(context.Background(), []contracts.Feed{feed1, feed2})

	assert.Nil(t, got)
	require.Error(t, err)
	assert.ErrorIs(t, err, errFirst)
	assert.ErrorIs(t, err, errSecond)
}
