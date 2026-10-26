package config

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

type mockCatwalkClient struct {
	providers []catwalk.Provider
	err       error
	callCount atomic.Int32
}

func (m *mockCatwalkClient) GetProviders(ctx context.Context, etag string) ([]catwalk.Provider, error) {
	m.callCount.Add(1)
	return m.providers, m.err
}

// snapshot waits for the background refresh to replace the in-memory result
// and returns the settled value.
func (s *catwalkSync) settled(t *testing.T) []catwalk.Provider {
	t.Helper()
	var out []catwalk.Provider
	s.mu.RLock()
	out = s.result
	s.mu.RUnlock()
	return out
}

func TestCatwalkSync_Init(t *testing.T) {
	t.Parallel()

	syncer := &catwalkSync{}
	client := &mockCatwalkClient{}
	path := "/tmp/test.json"

	syncer.Init(client, path, true)

	require.True(t, syncer.init.Load())
	require.Equal(t, client, syncer.client)
	require.Equal(t, path, syncer.cache.path)
	require.True(t, syncer.autoupdate)
}

func TestCatwalkSync_GetPanicIfNotInit(t *testing.T) {
	t.Parallel()

	syncer := &catwalkSync{}
	require.Panics(t, func() {
		_, _ = syncer.Get(t.Context())
	})
}

func TestCatwalkSync_GetWithAutoUpdateDisabled(t *testing.T) {
	t.Parallel()

	syncer := &catwalkSync{}
	client := &mockCatwalkClient{
		providers: []catwalk.Provider{{Name: "should-not-be-used"}},
	}
	path := t.TempDir() + "/providers.json"

	syncer.Init(client, path, false)

	providers, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, providers)
	require.Equal(t, int32(0), client.callCount.Load(), "Client should not be called when autoupdate is disabled")

	// Should return embedded providers.
	for _, p := range providers {
		require.NotEqual(t, "should-not-be-used", p.Name)
	}
}

func TestCatwalkSync_GetFreshProviders(t *testing.T) {
	t.Parallel()

	syncer := &catwalkSync{}
	client := &mockCatwalkClient{
		providers: []catwalk.Provider{
			{Name: "Fresh Provider", ID: "fresh"},
		},
	}
	path := t.TempDir() + "/providers.json"

	syncer.Init(client, path, true)

	// Get returns the embedded snapshot immediately without waiting on the
	// (mock) network refresh.
	providers, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, providers)

	// The background refresh replaces the result and writes the cache.
	require.Eventually(t, func() bool {
		result := syncer.settled(t)
		return len(result) == 1 && result[0].Name == "Fresh Provider"
	}, time.Second, 5*time.Millisecond)
	require.Eventually(t, func() bool {
		fileInfo, err := os.Stat(path)
		return err == nil && !fileInfo.IsDir()
	}, time.Second, 5*time.Millisecond)
	require.Equal(t, int32(1), client.callCount.Load())
}

func TestCatwalkSync_GetNotModifiedUsesCached(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := tmpDir + "/providers.json"

	// Create cache file.
	cachedProviders := []catwalk.Provider{
		{Name: "Cached Provider", ID: "cached"},
	}
	data, err := json.Marshal(cachedProviders)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))

	syncer := &catwalkSync{}
	client := &mockCatwalkClient{
		err: catwalk.ErrNotModified,
	}

	syncer.Init(client, path, true)

	providers, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.Len(t, providers, 1)
	require.Equal(t, "Cached Provider", providers[0].Name)

	require.Eventually(t, func() bool {
		return client.callCount.Load() == 1
	}, time.Second, 5*time.Millisecond)

	// ErrNotModified must leave the cached snapshot in place.
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, "Cached Provider", syncer.settled(t)[0].Name)
}

func TestCatwalkSync_GetEmptyResultKeepsCached(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := tmpDir + "/providers.json"

	// Create cache file.
	cachedProviders := []catwalk.Provider{
		{Name: "Cached Provider", ID: "cached"},
	}
	data, err := json.Marshal(cachedProviders)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))

	syncer := &catwalkSync{}
	client := &mockCatwalkClient{
		providers: []catwalk.Provider{}, // Empty result.
	}

	syncer.Init(client, path, true)

	providers, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.Len(t, providers, 1)
	require.Equal(t, "Cached Provider", providers[0].Name)

	require.Eventually(t, func() bool {
		return client.callCount.Load() == 1
	}, time.Second, 5*time.Millisecond)

	// An empty refresh must not clobber the cached result.
	time.Sleep(50 * time.Millisecond)
	require.Len(t, syncer.settled(t), 1)
	require.Equal(t, "Cached Provider", syncer.settled(t)[0].Name)
}

func TestCatwalkSync_GetEmptyCacheDefaultsToEmbedded(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := tmpDir + "/providers.json"

	// Create empty cache file.
	emptyProviders := []catwalk.Provider{}
	data, err := json.Marshal(emptyProviders)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))

	syncer := &catwalkSync{}
	client := &mockCatwalkClient{
		err: errors.New("network error"),
	}

	syncer.Init(client, path, true)

	providers, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, providers, "Should fall back to embedded providers")

	// Verify it's embedded providers by checking we have multiple common ones.
	require.Greater(t, len(providers), 5)

	require.Eventually(t, func() bool {
		return client.callCount.Load() == 1
	}, time.Second, 5*time.Millisecond)
}

func TestCatwalkSync_GetClientError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := tmpDir + "/providers.json"

	syncer := &catwalkSync{}
	client := &mockCatwalkClient{
		err: errors.New("network error"),
	}

	syncer.Init(client, path, true)

	providers, err := syncer.Get(t.Context())
	require.NoError(t, err) // Should fall back to embedded.
	require.NotEmpty(t, providers)

	require.Eventually(t, func() bool {
		return client.callCount.Load() == 1
	}, time.Second, 5*time.Millisecond)
}

func TestCatwalkSync_GetCalledMultipleTimesUsesOnce(t *testing.T) {
	t.Parallel()

	syncer := &catwalkSync{}
	client := &mockCatwalkClient{
		providers: []catwalk.Provider{
			{Name: "Provider", ID: "test"},
		},
	}
	path := t.TempDir() + "/providers.json"

	syncer.Init(client, path, true)

	// Call Get multiple times.
	providers1, err1 := syncer.Get(t.Context())
	require.NoError(t, err1)
	require.NotEmpty(t, providers1)

	providers2, err2 := syncer.Get(t.Context())
	require.NoError(t, err2)
	require.NotEmpty(t, providers2)

	// The refresh runs once in the background despite repeated Gets.
	require.Eventually(t, func() bool {
		return client.callCount.Load() == 1
	}, time.Second, 5*time.Millisecond)
	require.Eventually(t, func() bool {
		result := syncer.settled(t)
		return len(result) == 1 && result[0].Name == "Provider"
	}, time.Second, 5*time.Millisecond)
}
