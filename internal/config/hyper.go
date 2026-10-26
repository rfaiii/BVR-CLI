package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	xetag "github.com/charmbracelet/x/etag"
	"github.com/richavery/bvr-cli/internal/agent/hyper"
)

type hyperClient interface {
	Get(context.Context, string) (catwalk.Provider, error)
}

var _ syncer[catwalk.Provider] = (*hyperSync)(nil)

type hyperSync struct {
	mu         sync.RWMutex
	once       sync.Once
	result     catwalk.Provider
	cache      cache[catwalk.Provider]
	client     hyperClient
	autoupdate bool
	init       atomic.Bool

	// onLanded runs on the refresh goroutine after a newer result has been
	// stored; see catwalkSync.onLanded.
	onLanded func()
}

func (s *hyperSync) Init(client hyperClient, path string, autoupdate bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = client
	s.cache = newCache[catwalk.Provider](path)
	s.autoupdate = autoupdate
	s.init.Store(true)
}

// setOnLanded installs the callback for a landed refresh. See
// catwalkSync.setOnLanded.
func (s *hyperSync) setOnLanded(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onLanded = fn
}

func (s *hyperSync) Get(ctx context.Context) (catwalk.Provider, error) {
	if !s.init.Load() {
		panic("called Get before Init")
	}

	s.once.Do(func() {
		if !s.autoupdate {
			slog.Info("Using embedded Hyper provider")
			s.result = hyper.Embedded()
			return
		}

		cached, etag, cachedErr := s.cache.Get()
		if cached.ID == "" || cachedErr != nil {
			// if cached file is empty, default to embedded provider
			cached = hyper.Embedded()
		}
		s.result = cached

		// Refresh in the background so startup never blocks on the network.
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()

			slog.Info("Fetching Hyper provider in background")
			result, err := s.client.Get(ctx, etag)
			switch {
			case errors.Is(err, catwalk.ErrNotModified):
				return
			case errors.Is(err, context.DeadlineExceeded):
				slog.Warn("Hyper provider not updated in time")
				return
			case err != nil:
				slog.Warn("Hyper provider not updated", "error", err)
				return
			case len(result.Models) == 0:
				slog.Warn("Hyper did not return any models")
				return
			}

			s.mu.Lock()
			s.result = result
			landed := s.onLanded
			s.mu.Unlock()
			if err := s.cache.Store(result); err != nil {
				slog.Warn("Failed to store hyper cache", "error", err)
			}
			if landed != nil {
				landed()
			}
		}()
	})

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result, nil
}

// current returns the memoized result, reporting false when the syncer was
// never initialized. See catwalkSync.current.
func (s *hyperSync) current() (catwalk.Provider, bool) {
	if !s.init.Load() {
		return catwalk.Provider{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result, true
}

var _ hyperClient = realHyperClient{}

type realHyperClient struct {
	baseURL string
}

// Get implements hyperClient.
func (r realHyperClient) Get(ctx context.Context, etag string) (catwalk.Provider, error) {
	var result catwalk.Provider
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		r.baseURL+"/api/v1/provider",
		nil,
	)
	if err != nil {
		return result, fmt.Errorf("could not create request: %w", err)
	}
	xetag.Request(req, etag)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode == http.StatusNotModified {
		return result, catwalk.ErrNotModified
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
