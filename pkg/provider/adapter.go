package provider

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/feed-relay/rsscast"

	"github.com/feed-relay/localdir/internal/fs"
)

type Adapter interface {
	Feed(ctx context.Context, feed Feed) (*rsscast.Feed, error)
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

func NewAdapter(config Config) Adapter {
	return &adapter{
		config: config,
	}
}

func (a *adapter) Feed(ctx context.Context, feed Feed) (*rsscast.Feed, error) {
	baseURL := feed.Link()
	fallbackCaption := "Feed Relay Localdir"

	cover := ""
	if feed.Image() != "" {
		cover, _ = url.JoinPath(baseURL, feed.Image())
	}

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
				Length: audio.Length,
			},
		}
		item := rsscast.NewItem(itemData).
			WithPubDate(audio.ModTime).
			WithDescription(itemData.Title).
			WithItunesDuration(audio.Duration).
			WithLink(enclosureURL).
			WithItunesExplicit(rsscast.ExplicitFalse).
			WithItunesTitle(itemData.Title).
			WithItunesEpisodeType(rsscast.EpisodeFull).
			WithItunesAuthor(itemAuthor)

		itemDescription := strings.TrimSpace(audio.Metadata.Comment())
		if itemDescription != "" {
			item.WithDescription(itemDescription)
		}

		// TODO itemImage
		//itemImage := ""
		//if itemImage != "" {
		//	item.WithItunesImage(itemImage)
		//}

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
