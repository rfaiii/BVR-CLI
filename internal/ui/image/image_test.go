package image

import (
	"image"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestResetCache and TestResetIdempotent work against the package-global cache,
// so neither may call t.Parallel(): it lets one test's inserts land between the
// other's ResetCache() and its length check, which fails the assertion
// intermittently.
func TestResetCache(t *testing.T) {
	cachedMutex.Lock()
	cachedImages[imageKey{id: "a", cols: 10, rows: 10}] = cachedImage{
		img:  image.NewRGBA(image.Rect(0, 0, 1, 1)),
		cols: 10,
		rows: 10,
	}
	cachedImages[imageKey{id: "b", cols: 20, rows: 20}] = cachedImage{
		img:  image.NewRGBA(image.Rect(0, 0, 1, 1)),
		cols: 20,
		rows: 20,
	}
	cachedMutex.Unlock()

	ResetCache()

	cachedMutex.RLock()
	length := len(cachedImages)
	cachedMutex.RUnlock()

	require.Equal(t, 0, length)
}

func TestResetIdempotent(t *testing.T) {
	// Calling Reset on an empty cache should not panic.
	// Not parallel: see the note on TestResetCache.
	ResetCache()

	cachedMutex.RLock()
	length := len(cachedImages)
	cachedMutex.RUnlock()

	require.Equal(t, 0, length)
}
