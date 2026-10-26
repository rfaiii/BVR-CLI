package config

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

// The provider catalog is fetched in the background so the first frame never
// waits on the network, and Providers() memoizes what it served at startup.
// Those two facts combine into a bug the tests here exist to keep fixed: a
// refresh that lands later would update only the syncers, leaving every
// caller of Providers() — the model dialog above all — reading the startup
// catalog for the life of the process. So the landing has to rebuild the
// memoized list, and it has to say so.

// gatedCatwalkClient parks inside GetProviders until release is closed, which
// is how a test holds a refresh open and then lets it land on demand.
type gatedCatwalkClient struct {
	release   chan struct{}
	providers []catwalk.Provider
}

func (g *gatedCatwalkClient) GetProviders(ctx context.Context, _ string) ([]catwalk.Provider, error) {
	select {
	case <-g.release:
		return g.providers, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type gatedHyperClient struct {
	release  chan struct{}
	provider catwalk.Provider
}

func (g *gatedHyperClient) Get(ctx context.Context, _ string) (catwalk.Provider, error) {
	select {
	case <-g.release:
		return g.provider, nil
	case <-ctx.Done():
		return catwalk.Provider{}, ctx.Err()
	}
}

func providerIDs(list []catwalk.Provider) []string {
	ids := make([]string, 0, len(list))
	for _, p := range list {
		ids = append(ids, string(p.ID))
	}
	return ids
}

func containsID(list []catwalk.Provider, id string) bool {
	return slices.ContainsFunc(list, func(p catwalk.Provider) bool {
		return string(p.ID) == id
	})
}

// servingList reads the memoized list exactly the way Providers does once its
// once has run. Used by the tests here because calling Providers() itself would
// re-Init the syncers underneath a refresh that is still parked in a mock.
func servingList() []catwalk.Provider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	return providerList
}

// startParkedCatalogs installs gated syncers on the package globals, serves the
// cached catalog, and leaves the refresh parked mid-flight. It returns the two
// gates.
//
// The hook is installed here rather than by Providers() so the whole sequence
// stays race-free: Init happens before the refresh goroutine exists, and only
// onLanded is written afterwards, under the syncer's lock.
func startParkedCatalogs(t *testing.T, cached []catwalk.Provider, hyperCached catwalk.Provider) (
	releaseCatwalk, releaseHyper func(),
) {
	t.Helper()

	catwalkPath := t.TempDir() + "/providers.json"
	hyperPath := t.TempDir() + "/hyper.json"
	require.NoError(t, newCache[[]catwalk.Provider](catwalkPath).Store(cached))
	require.NoError(t, newCache[catwalk.Provider](hyperPath).Store(hyperCached))

	catwalkClient := &gatedCatwalkClient{
		release:   make(chan struct{}),
		providers: []catwalk.Provider{{ID: "fresh-provider", Name: "Fresh"}},
	}
	hyperClient := &gatedHyperClient{
		release:  make(chan struct{}),
		provider: catwalk.Provider{ID: "hyper", Name: "Fresh Hyper", Models: []catwalk.Model{{ID: "fresh-model"}}},
	}

	catwalkSyncer.Init(catwalkClient, catwalkPath, true)
	hyperSyncer.Init(hyperClient, hyperPath, true)
	catwalkSyncer.setOnLanded(onProvidersLanded)
	hyperSyncer.setOnLanded(onProvidersLanded)

	ctx := t.Context()
	_, err := catwalkSyncer.Get(ctx)
	require.NoError(t, err)
	_, err = hyperSyncer.Get(ctx)
	require.NoError(t, err)

	// Unblock anything still parked when the test ends, or the goroutine
	// outlives it.
	var onceCatwalk, onceHyper sync.Once
	t.Cleanup(func() {
		onceCatwalk.Do(func() { close(catwalkClient.release) })
		onceHyper.Do(func() { close(hyperClient.release) })
	})
	return func() { onceCatwalk.Do(func() { close(catwalkClient.release) }) },
		func() { onceHyper.Do(func() { close(hyperClient.release) }) }
}

func TestLandingRebuildsTheMemoizedProviderList(t *testing.T) {
	resetProviderState()
	defer resetProviderState()

	releaseCatwalk, releaseHyper := startParkedCatalogs(t,
		[]catwalk.Provider{{ID: "startup-provider", Name: "Startup"}},
		catwalk.Provider{ID: "hyper", Name: "Startup Hyper"},
	)

	// Stand in for the list Providers() builds and memoizes at startup.
	providerMu.Lock()
	providerList = currentProviderList()
	providerMu.Unlock()

	startup := servingList()
	require.True(t, containsID(startup, "startup-provider"), "precondition: the cached catalog is what is served")
	require.False(t, containsID(startup, "fresh-provider"))
	require.Equal(t, uint64(0), ProviderVersion(), "nothing has landed yet")

	// Hyper first, so the generation count below is deterministic.
	releaseHyper()
	require.Eventually(t, func() bool {
		return ProviderVersion() == 1
	}, 2*time.Second, 5*time.Millisecond, "each landed catalog should advance the generation by one")

	releaseCatwalk()

	require.Eventually(t, func() bool {
		return containsID(servingList(), "fresh-provider")
	}, 2*time.Second, 5*time.Millisecond,
		"a landed catalog must replace the memoized list, not just the syncer")
	require.Equal(t, uint64(2), ProviderVersion())

	// The rebuild swaps in a new slice rather than editing the one already
	// handed out, so a reader holding the startup list never sees it change
	// underneath it.
	require.True(t, containsID(startup, "startup-provider"))
	require.False(t, containsID(startup, "fresh-provider"),
		"a list already handed to a caller must not be mutated in place")

	// The rebuild is the shared assembly, so Hyper and Cline still lead it.
	reloaded := servingList()
	require.Equal(t, []string{"hyper", string(ClineProviderID), "fresh-provider"}, providerIDs(reloaded))
}

func TestLandingNotifiesSubscribers(t *testing.T) {
	resetProviderState()
	defer resetProviderState()

	releaseCatwalk, releaseHyper := startParkedCatalogs(t,
		[]catwalk.Provider{{ID: "startup-provider", Name: "Startup"}},
		catwalk.Provider{ID: "hyper", Name: "Startup Hyper"},
	)
	defer releaseCatwalk()
	defer releaseHyper()

	providerMu.Lock()
	providerList = currentProviderList()
	providerMu.Unlock()

	notified := make(chan uint64, 4)
	OnProvidersUpdated(func(generation uint64) { notified <- generation })

	releaseCatwalk()

	select {
	case generation := <-notified:
		require.Equal(t, ProviderVersion(), generation,
			"a notification must carry the generation the list now matches")
	case <-time.After(2 * time.Second):
		t.Fatal("a landed catalog did not notify subscribers")
	}

	// Both catalogs feed one list, so each landing is its own notification.
	releaseHyper()
	select {
	case generation := <-notified:
		require.Equal(t, uint64(2), generation)
	case <-time.After(2 * time.Second):
		t.Fatal("the second catalog did not notify subscribers")
	}
	require.Eventually(t, func() bool {
		return containsID(servingList(), "fresh-provider") &&
			len(servingList()) == 3
	}, 2*time.Second, 5*time.Millisecond,
		"the list must settle on both landed catalogs")
}

func TestOnProvidersUpdatedReplaysALandingThatAlreadyHappened(t *testing.T) {
	resetProviderState()
	defer resetProviderState()

	// A refresh is one network round trip and can finish before the TUI exists
	// to subscribe. Without a replay nothing would ever tell the UI that the
	// catalog it read at startup had been replaced.
	onProvidersLanded()
	require.Equal(t, uint64(1), ProviderVersion())

	var mu sync.Mutex
	var replayed []uint64
	OnProvidersUpdated(func(generation uint64) {
		mu.Lock()
		replayed = append(replayed, generation)
		mu.Unlock()
	})

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []uint64{1}, replayed,
		"registering after a landing must report the landing that was missed")
}

func TestLandingWithNoCatalogsServesOnlyCline(t *testing.T) {
	resetProviderState()
	defer resetProviderState()

	// DisableDefaultProviders skips both catalogs, so neither syncer is ever
	// initialized and there is nothing to rebuild. The list must stay the
	// custom-providers-only shape it was at startup.
	cfg := &Config{Options: &Options{DisableDefaultProviders: true}}
	providers, err := Providers(cfg)
	require.NoError(t, err)
	require.Equal(t, []string{string(ClineProviderID)}, providerIDs(providers))

	onProvidersLanded()
	require.Equal(t, []string{string(ClineProviderID)}, providerIDs(servingList()),
		"a landing must not invent catalog entries when no catalog was consulted")
}

func TestAssembleProvidersKeepsHyperAndClineLeading(t *testing.T) {
	items := []catwalk.Provider{{ID: "catwalk-a"}, {ID: "catwalk-b"}}
	hyperProvider := catwalk.Provider{ID: "hyper"}

	list := assembleProviders(items, hyperProvider, true)
	require.Equal(t, []string{"hyper", string(ClineProviderID), "catwalk-a", "catwalk-b"}, providerIDs(list),
		"the rebuilt list must be assembled exactly like the startup list")

	noHyper := assembleProviders(items, catwalk.Provider{}, false)
	require.Equal(t, []string{string(ClineProviderID), "catwalk-a", "catwalk-b"}, providerIDs(noHyper))

	require.Len(t, items, 2, "assembleProviders must not append into the caller's slice")
}
