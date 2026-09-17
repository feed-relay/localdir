package provider

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/feed-relay/contracts"
	"github.com/feed-relay/rsscast"

	"github.com/feed-relay/localdir/internal/fs"
)

//go:generate moq --out ./mocks/adapter_mock.go --pkg mocks --skip-ensure --with-resets -fmt goimports . Adapter
//go:generate moq --out ./mocks/config_mock.go --pkg mocks --skip-ensure --with-resets -fmt goimports . Config

type Adapter interface {
	Feed(ctx context.Context, feed contracts.Feed) (*rsscast.Feed, error)
}

// Config supplies feed-level metadata.
type Config interface {
	Generator() string
	ItunesOwnerName() string
	ItunesOwnerEmail() string
}

type adapter struct {
	config Config
}

func (a *adapter) Feed(_ context.Context, f contracts.Feed) (*rsscast.Feed, error) {
	feed, ok := f.(Feed)
	if !ok {
		return nil, fmt.Errorf("unsupported feed type: %T", f)
	}
	baseURL := feed.Link()
	fallbackCaption := "Feed Relay Localdir"

	cover, _ := url.JoinPath(baseURL, feed.Image())

	title := feed.Title()
	if title == "" {
		title = fallbackCaption
	}

	description := feed.Description()
	if description == "" {
		description = fallbackCaption
	}

	feedData := rsscast.FeedData{
		Title:       title,
		Description: description,
		Image:       cover,
		Language:    "ru",
		Explicit:    rsscast.ExplicitFalse,
		Categories: []rsscast.Category{
			rsscast.NewCategory("Society & Culture"),
		},
	}
	rssFeed := rsscast.NewFeed(feedData).
		WithAuthor(feedData.Title).
		WithLink(baseURL).
		WithPubDate(time.Now()).
		WithLastBuildDate(time.Now()).
		WithItunesTitle(feedData.Title).
		WithGenerator(a.config.Generator()).
		WithItunesSummary(feedData.Description).
		WithItunesOwner(a.config.ItunesOwnerName(), a.config.ItunesOwnerEmail()).
		WithItunesType(rsscast.TypeEpisodic)

	audios := feed.AudioFiles()
	for _, audio := range audios {
		enclosureLength := audio.Length
		if enclosureLength <= 0 {
			enclosureLength = 1_000_000 /// 1Mb
		}
		enclosureType, _ := enclosureTypeByAudioType(audio.AudioType)
		enclosureURL, _ := url.JoinPath(baseURL, audio.Name)

		itemTitle := strings.TrimSpace(audio.Metadata.Title())
		if itemTitle == "" {
			itemTitle = audio.Name
		}

		itemAuthor := strings.TrimSpace(audio.Metadata.Artist())
		if itemAuthor == "" {
			itemAuthor = strings.TrimSpace(audio.Metadata.AlbumArtist())
		}
		if itemAuthor == "" {
			itemAuthor = fallbackCaption
		}

		itemData := rsscast.ItemData{
			Title: itemTitle,
			Guid:  audio.Name,
			Enclosure: rsscast.Enclosure{
				URL:    enclosureURL,
				Type:   enclosureType,
				Length: enclosureLength,
			},
		}
		item := rsscast.NewItem(itemData).
			WithPubDate(audio.ModTime).
			WithDescription(itemData.Title).
			WithItunesDuration(int64(audio.Duration.Seconds())).
			WithLink(enclosureURL).
			WithItunesExplicit(rsscast.ExplicitFalse).
			WithItunesTitle(itemData.Title).
			WithItunesEpisodeType(rsscast.EpisodeFull).
			WithItunesAuthor(itemAuthor)

		itemDescription := strings.TrimSpace(audio.Metadata.Comment())
		if itemDescription != "" {
			item.WithDescription(itemDescription)
		}

		var itemImage string
		pic := audio.Metadata.Picture()
		if pic != nil {
			itemImage, _ = picture(feed.Dir(), strings.TrimSuffix(audio.Name, filepath.Ext(audio.Name)), pic)
			if itemImage != "" {
				itemImage, _ = url.JoinPath(baseURL, itemImage)
			}
		}
		if itemImage != "" {
			item.WithItunesImage(itemImage)
		}

		rssFeed.AddItem(item)
	}

	return rssFeed, nil
}

func enclosureTypeByAudioType(audioType fs.AudioType) (rsscast.EnclosureType, bool) {
	switch audioType {
	case fs.Mp3:
		return rsscast.Mp3, true
	case fs.M4a:
		return rsscast.M4a, true
	default:
		return rsscast.Mp3, false
	}
}
