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
	"github.com/richavery/bvr-cli/internal/agent/hyper"
	"github.com/stretchr/testify/require"
)

type mockHyperClient struct {
	provider  catwalk.Provider
	err       error
	callCount atomic.Int32
}

func (m *mockHyperClient) Get(ctx context.Context, etag string) (catwalk.Provider, error) {
	m.callCount.Add(1)
	return m.provider, m.err
}

func (s *hyperSync) settled(t *testing.T) catwalk.Provider {
	t.Helper()
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result
}

func TestHyperSync_Init(t *testing.T) {
	t.Parallel()

	syncer := &hyperSync{}
	client := &mockHyperClient{}
	path := "/tmp/hyper.json"

	syncer.Init(client, path, true)

	require.True(t, syncer.init.Load())
	require.Equal(t, client, syncer.client)
	require.Equal(t, path, syncer.cache.path)
}

func TestHyperSync_GetPanicIfNotInit(t *testing.T) {
	t.Parallel()

	syncer := &hyperSync{}
	require.Panics(t, func() {
		_, _ = syncer.Get(t.Context())
	})
}

func TestHyperSync_GetFreshProvider(t *testing.T) {
	t.Parallel()

	syncer := &hyperSync{}
	client := &mockHyperClient{
		provider: catwalk.Provider{
			Name: "Hyper",
			ID:   "hyper",
			Models: []catwalk.Model{
				{ID: "model-1", Name: "Model 1"},
			},
		},
	}
	path := t.TempDir() + "/hyper.json"

	syncer.Init(client, path, true)

	// First Get returns the embedded snapshot without waiting on refresh.
	provider, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.Equal(t, hyper.Embedded().Name, provider.Name)

	require.Eventually(t, func() bool {
		return syncer.settled(t).Name == "Hyper"
	}, time.Second, 5*time.Millisecond)
	require.Eventually(t, func() bool {
		fileInfo, err := os.Stat(path)
		return err == nil && !fileInfo.IsDir()
	}, time.Second, 5*time.Millisecond)
	require.Equal(t, int32(1), client.callCount.Load())
}

func TestHyperSync_GetNotModifiedUsesCached(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := tmpDir + "/hyper.json"

	// Create cache file.
	cachedProvider := catwalk.Provider{
		Name: "Cached Hyper",
		ID:   "hyper",
	}
	data, err := json.Marshal(cachedProvider)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))

	syncer := &hyperSync{}
	client := &mockHyperClient{
		err: catwalk.ErrNotModified,
	}

	syncer.Init(client, path, true)

	provider, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.Equal(t, "Cached Hyper", provider.Name)

	require.Eventually(t, func() bool {
		return client.callCount.Load() == 1
	}, time.Second, 5*time.Millisecond)

	time.Sleep(50 * time.Millisecond)
	require.Equal(t, "Cached Hyper", syncer.settled(t).Name)
}

func TestHyperSync_GetClientError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := tmpDir + "/hyper.json"

	syncer := &hyperSync{}
	client := &mockHyperClient{
		err: errors.New("network error"),
	}

	syncer.Init(client, path, true)

	provider, err := syncer.Get(t.Context())
	require.NoError(t, err) // Should fall back to embedded.
	require.Equal(t, "Charm Hyper", provider.Name)
	require.Equal(t, catwalk.InferenceProvider("hyper"), provider.ID)

	require.Eventually(t, func() bool {
		return client.callCount.Load() == 1
	}, time.Second, 5*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, "Charm Hyper", syncer.settled(t).Name)
}

func TestHyperSync_GetEmptyCache(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := tmpDir + "/hyper.json"

	syncer := &hyperSync{}
	client := &mockHyperClient{
		provider: catwalk.Provider{
			Name: "Fresh Hyper",
			ID:   "hyper",
			Models: []catwalk.Model{
				{ID: "model-1", Name: "Model 1"},
			},
		},
	}

	syncer.Init(client, path, true)

	provider, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.Equal(t, hyper.Embedded().Name, provider.Name)

	require.Eventually(t, func() bool {
		return syncer.settled(t).Name == "Fresh Hyper"
	}, time.Second, 5*time.Millisecond)
}

func TestHyperSync_GetCalledMultipleTimesUsesOnce(t *testing.T) {
	t.Parallel()

	syncer := &hyperSync{}
	client := &mockHyperClient{
		provider: catwalk.Provider{
			Name: "Hyper",
			ID:   "hyper",
			Models: []catwalk.Model{
				{ID: "model-1", Name: "Model 1"},
			},
		},
	}
	path := t.TempDir() + "/hyper.json"

	syncer.Init(client, path, true)

	// Call Get multiple times.
	provider1, err1 := syncer.Get(t.Context())
	require.NoError(t, err1)
	require.NotEmpty(t, provider1.Name)

	provider2, err2 := syncer.Get(t.Context())
	require.NoError(t, err2)
	require.NotEmpty(t, provider2.Name)

	// The refresh runs once in the background despite repeated Gets.
	require.Eventually(t, func() bool {
		return client.callCount.Load() == 1
	}, time.Second, 5*time.Millisecond)
	require.Eventually(t, func() bool {
		return syncer.settled(t).Name == "Hyper"
	}, time.Second, 5*time.Millisecond)
}

func TestHyperSync_GetCacheStoreError(t *testing.T) {
	t.Parallel()

	// Create a file where we want a directory, causing mkdir to fail.
	tmpDir := t.TempDir()
	blockingFile := tmpDir + "/blocking"
	require.NoError(t, os.WriteFile(blockingFile, []byte("block"), 0o644))

	// Try to create cache in a subdirectory under the blocking file.
	path := blockingFile + "/subdir/hyper.json"

	syncer := &hyperSync{}
	client := &mockHyperClient{
		provider: catwalk.Provider{
			Name: "Hyper",
			ID:   "hyper",
			Models: []catwalk.Model{
				{ID: "model-1", Name: "Model 1"},
			},
		},
	}

	syncer.Init(client, path, true)

	provider, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.Equal(t, hyper.Embedded().Name, provider.Name)

	// A failed cache Store is logged only; the in-memory result still updates.
	require.Eventually(t, func() bool {
		return syncer.settled(t).Name == "Hyper"
	}, time.Second, 5*time.Millisecond)
}
