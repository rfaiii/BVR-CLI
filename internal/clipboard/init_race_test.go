package clipboard

import (
	"sync"
	"testing"
)

// app.New calls Init from a goroutine, and creating several workspaces starts
// several of them at once. On Linux that concurrent Init wrote the detected
// wl-clipboard backend into package state while other goroutines read it,
// which go test -race caught on the backend suite.
func TestInitIsSafeToCallConcurrently(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = Init()
			canWrite()
			canRead()
		}()
	}
	wg.Wait()
}
