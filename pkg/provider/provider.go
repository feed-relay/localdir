package provider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/feed-relay/contracts"
	"github.com/feed-relay/rsscast"

	"github.com/feed-relay/localdir/internal/fs"
)

//go:generate moq --out ./mocks/feed_mock.go --pkg mocks --skip-ensure --with-resets -fmt goimports . Feed

const Platform = "localdir"

const feedWorkers = 8

// Provider Local data.
//
// Should implement `contracts.Provider` interface
type Provider struct {
	finder  *fs.Finder
	adapter Adapter
}

func NewProvider(config Config, finder *fs.Finder) *Provider {
	return &Provider{
		finder:  finder,
		adapter: &adapter{config: config},
	}
}

// Feed is the subset of contracts.Feed that Feeds/feed need.
type Feed interface {
	contracts.Feed

	Dir() string
	AudioFiles() []fs.AudioFile
	SetFinder(finder *fs.Finder)
}

// feedTask is one feed to be processed by a worker; all of its
// shows are merged into a single feed.
type feedTask struct {
	feed Feed
}

func (p *Provider) Platform() string {
	return Platform
}

func (p *Provider) Feeds(ctx context.Context, feeds []contracts.Feed) (map[string]*rsscast.Feed, error) {
	var ff []Feed
	for _, f := range feeds {
		pf, ok := f.(Feed)
		if !ok {
			slog.Error("unsupported feed type", slog.Any("feed type", fmt.Errorf("%T", f)))
			continue
		}

		pf.SetFinder(p.finder)
		if len(pf.Shows()) == 0 {
			continue
		}
		ff = append(ff, pf)
	}

	return p.feeds(ctx, ff)
}

func (p *Provider) feeds(ctx context.Context, feeds []Feed) (map[string]*rsscast.Feed, error) {
	var allTasks []feedTask
	for _, f := range feeds {
		allTasks = append(allTasks, feedTask{feed: f})
	}

	workers := min(feedWorkers, len(allTasks))
	if workers == 0 {
		return nil, errors.New("localdir: no shows")
	}

	tasks := make(chan feedTask)

	var wg sync.WaitGroup
	var mu sync.Mutex

	rssFeeds := make(map[string]*rsscast.Feed)
	var errs []error

	for range workers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for task := range tasks {
				slug, feed, err := p.feed(ctx, task.feed)
				if err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
				if feed == nil {
					continue
				}

				mu.Lock()
				rssFeeds[slug] = feed
				mu.Unlock()
			}
		}()
	}

	for _, task := range allTasks {
		tasks <- task
	}

	close(tasks)
	wg.Wait()

	var err error
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}

	if len(rssFeeds) > 0 {
		return rssFeeds, err
	}

	return nil, err
}

func (p *Provider) feed(ctx context.Context, feed Feed) (string, *rsscast.Feed, error) {
	rssFeed, err := p.adapter.Feed(ctx, feed)
	if err != nil {
		return "", rssFeed, fmt.Errorf("build feed failed: %w", err)
	}

	slug := fmt.Sprintf("%s-%s", p.Platform(), feed.Slug())

	return slug, rssFeed, nil
}
