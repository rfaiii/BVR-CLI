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
}

func (s *catwalkSync) Init(client catwalkClient, path string, autoupdate bool) {
	s.client = client
	s.cache = newCache[[]catwalk.Provider](path)
	s.autoupdate = autoupdate
	s.init.Store(true)
}

func (s *catwalkSync) Get(ctx context.Context) ([]catwalk.Provider, error) {
	if !s.init.Load() {
		panic("called Get before Init")
	}

	s.once.Do(func() {
		if !s.autoupdate {
			slog.Info("Using embedded Catwalk providers")
			s.result = embedded.GetAll()
			return
		}

		cached, etag, cachedErr := s.cache.Get()
		if len(cached) == 0 || cachedErr != nil {
			// if cached file is empty, default to embedded providers
			cached = embedded.GetAll()
		}
		s.result = cached

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
			s.mu.Unlock()
			if err := s.cache.Store(result); err != nil {
				slog.Warn("Failed to store catwalk provider cache", "error", err)
			}
		}()
	})

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result, nil
}
