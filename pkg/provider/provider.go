package provider

import (
	"context"

	"github.com/feed-relay/contracts"
	"github.com/feed-relay/rsscast"
)

const Platform = "localdir"

// Provider Local data.
//
// Should implement `contracts.Provider` interface
type Provider struct {
	adapter Adapter
}

func (p *Provider) Platform() string {
	return Platform
}

func (p *Provider) Feeds(ctx context.Context, feeds []contracts.Feed) (map[string]*rsscast.Feed, error) {
	return nil, nil
}
