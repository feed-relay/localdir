package provider

import (
	"context"

	"github.com/feed-relay/contracts"
	"github.com/feed-relay/rsscast"
)

type Adapter interface {
	Feed(ctx context.Context, feed contracts.Feed) (*rsscast.Feed, error)
}
