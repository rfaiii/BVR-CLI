package config

import (
	"context"
	"testing"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

// Both syncers serve the on-disk cache and refresh over the network from a
// goroutine, which is what keeps the 45-second provider timeout off the
// startup path. Nothing here proved that: every existing mock returns
// immediately, so a regression back to a blocking Get would still pass.
// These clients never answer.

type blockedCatwalkClient struct{}

func (blockedCatwalkClient) GetProviders(context.Context, string) ([]catwalk.Provider, error) {
	<-make(chan struct{}) // block forever
	return nil, nil
}

type blockedHyperClient struct{}

func (blockedHyperClient) Get(context.Context, string) (catwalk.Provider, error) {
	<-make(chan struct{}) // block forever
	return catwalk.Provider{}, nil
}

func TestCatwalkSyncServesCacheWithoutTheNetwork(t *testing.T) {
	t.Parallel()

	cached := []catwalk.Provider{{ID: "cached-provider", Name: "cached"}}
	path := t.TempDir() + "/providers.json"
	require.NoError(t, newCache[[]catwalk.Provider](path).Store(cached))

	syncer := &catwalkSync{}
	syncer.Init(blockedCatwalkClient{}, path, true)

	started := time.Now()
	providers, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.Equal(t, cached, providers, "the cache should be served as-is")
	require.Less(t, time.Since(started), 500*time.Millisecond,
		"Get waited on the network; startup is behind the 45s provider timeout again")
}

func TestHyperSyncServesCacheWithoutTheNetwork(t *testing.T) {
	t.Parallel()

	cached := catwalk.Provider{ID: "hyper", Name: "cached-hyper"}
	path := t.TempDir() + "/hyper.json"
	require.NoError(t, newCache[catwalk.Provider](path).Store(cached))

	syncer := &hyperSync{}
	syncer.Init(blockedHyperClient{}, path, true)

	started := time.Now()
	provider, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.Equal(t, cached, provider, "the cache should be served as-is")
	require.Less(t, time.Since(started), 500*time.Millisecond,
		"Get waited on the network; startup is behind the provider timeout again")
}
