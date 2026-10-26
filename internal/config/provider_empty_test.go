package config

import (
	"context"
	"os"
	"testing"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	"github.com/stretchr/testify/require"
)

type emptyProviderClient struct{}

func (m *emptyProviderClient) GetProviders(context.Context, string) ([]catwalk.Provider, error) {
	return []catwalk.Provider{}, nil
}

// TestCatwalkSync_GetEmptyResultFromClient tests that when the background
// refresh returns an empty list, the embedded fallback stays in place and no
// cache file is written.
func TestCatwalkSync_GetEmptyResultFromClient(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := tmpDir + "/providers.json"

	syncer := &catwalkSync{}
	client := &emptyProviderClient{}

	syncer.Init(client, path, true)

	providers, err := syncer.Get(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, providers) // Embedded providers as fallback.

	// Give the background refresh time to (not) write anything.
	time.Sleep(100 * time.Millisecond)
	_, statErr := os.Stat(path)
	require.True(t, os.IsNotExist(statErr), "Cache file should not exist for empty results")
	require.NotEmpty(t, syncer.settled(t))
}
