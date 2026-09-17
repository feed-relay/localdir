package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/meesooqa/go-cfg"
	"github.com/meesooqa/go-lgr"

	"github.com/feed-relay/contracts"

	"github.com/feed-relay/localdir/internal/config"
	"github.com/feed-relay/localdir/internal/fs"
	"github.com/feed-relay/localdir/internal/media"
	"github.com/feed-relay/localdir/internal/writer"
	"github.com/feed-relay/localdir/pkg/provider"
)

func main() {
	conf, err := cfg.Load[config.AppConfig]("etc/config.yml")
	if err != nil {
		log.Fatal(err)
	}
	logger, err := lgr.New(conf.Logger)
	if err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(logger)

	feeds := make([]contracts.Feed, len(conf.Feeds))
	feedsOutputDir := make(map[string]string, len(conf.Feeds))
	for i, feed := range conf.Feeds {
		feeds[i] = &feed
		fullSlug := fmt.Sprintf("%s-%s", provider.Platform, feed.Slug())
		feedsOutputDir[fullSlug] = feed.Dir()
	}

	finder := fs.NewFinder(media.NewMetadataReader(), media.NewFfprobeDurationReader(&media.ExecCommandRunner{}))
	p := provider.NewProvider(conf, finder)
	rssFeeds, err := p.Feeds(context.Background(), feeds)
	if err != nil {
		log.Fatal(err)
	}

	w := writer.NewXmlFileWriter()
	for slug, feed := range rssFeeds {
		err = w.Write(feedsOutputDir[slug], slug, feed)
		if err != nil {
			slog.Error("write", slog.Any("err", err))
			continue
		}
	}
}
