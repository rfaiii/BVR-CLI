package config

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/catwalk/pkg/embedded"
)

type catwalkClient interface {
	GetProviders(context.Context, string) ([]catwalk.Provider, error)
}

var _ syncer[[]catwalk.Provider] = (*catwalkSync)(nil)

type catwalkSync struct {
	mu         sync.RWMutex
	once       sync.Once
	result     []catwalk.Provider
	cache      cache[[]catwalk.Provider]
	client     catwalkClient
	autoupdate bool
	init       atomic.Bool

	// onLanded runs on the refresh goroutine after a newer result has been
	// stored. Providers installs it on the package syncers so the memoized
	// provider list gets rebuilt; a syncer built by a test leaves it nil and
	// its refresh stays local.
	onLanded func()
}

func (s *catwalkSync) Init(client catwalkClient, path string, autoupdate bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = client
	s.cache = newCache[[]catwalk.Provider](path)
	s.autoupdate = autoupdate
	s.init.Store(true)
}

// setOnLanded installs the callback for a landed refresh. It is a method
// rather than a field write because the refresh goroutine reads the callback
// under mu, and a caller may install it after that goroutine is already
// parked in the client.
func (s *catwalkSync) setOnLanded(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onLanded = fn
}

func (s *catwalkSync) Get(ctx context.Context) ([]catwalk.Provider, error) {
	if !s.init.Load() {
		panic("called Get before Init")
	}

	s.once.Do(func() {
		if !s.autoupdate {
			slog.Info("Using embedded Catwalk providers")
			// current() reads result under mu. The other catalog's refresh
			// can call that while this Once is still storing the embedded list.
			s.mu.Lock()
			s.result = embedded.GetAll()
			s.mu.Unlock()
			return
		}

		cached, etag, cachedErr := s.cache.Get()
		if len(cached) == 0 || cachedErr != nil {
			// if cached file is empty, default to embedded providers
			cached = embedded.GetAll()
		}
		s.mu.Lock()
		s.result = cached
		s.mu.Unlock()

		// Refresh in the background so callers never block on the network;
		// the fresh list lands in the cache for the next start and in memory
		// for any later read in this process. Deliberately decoupled from the
		// caller's ctx, which may already be canceled.
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()

			slog.Info("Fetching providers from Catwalk in background")
			result, err := s.client.GetProviders(ctx, etag)
			switch {
			case errors.Is(err, catwalk.ErrNotModified):
				return
			case errors.Is(err, context.DeadlineExceeded):
				slog.Warn("Catwalk providers not updated in time")
				return
			case err != nil:
				slog.Warn("Catwalk providers not updated", "error", err)
				return
			case len(result) == 0:
				slog.Warn("Empty providers list from catwalk")
				return
			}

			s.mu.Lock()
			s.result = result
			landed := s.onLanded
			s.mu.Unlock()
			if err := s.cache.Store(result); err != nil {
				slog.Warn("Failed to store catwalk provider cache", "error", err)
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
// never initialized. Unlike Get it cannot start a refresh, which is what
// makes it safe for rebuilding the provider list after a landing.
func (s *catwalkSync) current() ([]catwalk.Provider, bool) {
	if !s.init.Load() {
		return nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result, true
}
