package config

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/catwalk/pkg/embedded"
	"github.com/charmbracelet/x/etag"
	"github.com/richavery/bvr-cli/internal/agent/hyper"
	"github.com/richavery/bvr-cli/internal/home"
)

type syncer[T any] interface {
	Get(context.Context) (T, error)
}

var (
	providerOnce sync.Once

	// providerMu guards providerList because a background catalog refresh
	// replaces it while readers hold the slice it returned earlier. Swaps
	// publish a brand-new slice and no Provider is ever mutated in place,
	// so a reader keeps seeing a consistent (if slightly stale) list.
	providerMu   sync.RWMutex
	providerList []catwalk.Provider
	providerErr  error

	// syncerMu guards the package syncer pointers. Tests replace them
	// while a background refresh may still be reading them.
	syncerMu sync.RWMutex

	// providerVersion counts catalog landings. It is bumped after the list
	// is swapped, so a consumer that observes version N knows the list
	// matching N is already served. A landing that happens before anything
	// has subscribed would otherwise be missed for the life of the process,
	// which is what OnProvidersUpdated's replay is for.
	providerVersion atomic.Uint64

	// providerHooksMu guards the listener list and the "a landing has
	// already happened" flag. Hooks run on the refresh goroutine and must
	// not block.
	providerHooksMu sync.Mutex
	providerHooks   []func(version uint64)
)

// file to cache provider data
func cachePathFor(name string) string {
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome != "" {
		return filepath.Join(xdgDataHome, appName, name+".json")
	}

	// return the path to the main data directory
	// for windows, it should be in `%LOCALAPPDATA%/bvr/`
	// for linux and macOS, it should be in `$HOME/.local/share/bvr/`
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(localAppData, appName, name+".json")
	}

	return filepath.Join(home.Dir(), ".local", "share", appName, name+".json")
}

// UpdateProviders updates the Catwalk providers list from a specified source.
func UpdateProviders(pathOrURL string) error {
	var providers []catwalk.Provider
	pathOrURL = cmp.Or(pathOrURL, os.Getenv("CATWALK_URL"), defaultCatwalkURL)

	switch {
	case pathOrURL == "embedded":
		providers = embedded.GetAll()
	case strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://"):
		var err error
		providers, err = catwalk.NewWithURL(pathOrURL).GetProviders(context.Background(), "")
		if err != nil {
			return fmt.Errorf("failed to fetch providers from Catwalk: %w", err)
		}
	default:
		content, err := os.ReadFile(pathOrURL)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if err := json.Unmarshal(content, &providers); err != nil {
			return fmt.Errorf("failed to unmarshal provider data: %w", err)
		}
		if len(providers) == 0 {
			return fmt.Errorf("no providers found in the provided source")
		}
	}

	if err := newCache[[]catwalk.Provider](cachePathFor("providers")).Store(providers); err != nil {
		return fmt.Errorf("failed to save providers to cache: %w", err)
	}

	slog.Info("Providers updated successfully", "count", len(providers), "from", pathOrURL, "to", cachePathFor)
	return nil
}

// UpdateHyper updates the Hyper provider information from a specified URL.
func UpdateHyper(pathOrURL string) error {
	var provider catwalk.Provider
	pathOrURL = cmp.Or(pathOrURL, hyper.BaseURL())

	switch {
	case pathOrURL == "embedded":
		provider = hyper.Embedded()
	case strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://"):
		client := realHyperClient{baseURL: pathOrURL}
		var err error
		provider, err = client.Get(context.Background(), "")
		if err != nil {
			return fmt.Errorf("failed to fetch provider from Hyper: %w", err)
		}
	default:
		content, err := os.ReadFile(pathOrURL)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if err := json.Unmarshal(content, &provider); err != nil {
			return fmt.Errorf("failed to unmarshal provider data: %w", err)
		}
	}

	if err := newCache[catwalk.Provider](cachePathFor("hyper")).Store(provider); err != nil {
		return fmt.Errorf("failed to save Hyper provider to cache: %w", err)
	}

	slog.Info("Hyper provider updated successfully", "from", pathOrURL, "to", cachePathFor("hyper"))
	return nil
}

var (
	catwalkSyncer = &catwalkSync{}
	hyperSyncer   = &hyperSync{}
)

// Providers returns the list of providers, taking into account cached results
// and whether or not auto update is enabled.
//
// It will:
// 1. if auto update is disabled, it'll return the embedded providers at the
// time of release.
// 2. load the cached providers
// 3. try to get the fresh list of providers, and return either this new list,
// the cached list, or the embedded list if all others fail.
//
// The result is memoized but not frozen: when a background refresh lands a
// newer catalog the list is rebuilt from the same sources, so a long-running
// process serves the fresh providers instead of the ones read at startup.
// [OnProvidersUpdated] is how listeners hear about that.
func Providers(cfg *Config) ([]catwalk.Provider, error) {
	providerOnce.Do(func() {
		autoupdate := !cfg.Options.DisableProviderAutoUpdate
		customProvidersOnly := cfg.Options.DisableDefaultProviders

		var providerErrs []error
		if !customProvidersOnly {
			var wg sync.WaitGroup
			var catwalkErr, hyperErr error

			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()

			// Copy the pointers under the lock, then use those values. A test
			// can replace the globals while these fetches are in flight.
			syncerMu.RLock()
			catwalkCatalog := catwalkSyncer
			hyperCatalog := hyperSyncer
			syncerMu.RUnlock()

			// Install the hooks before the syncers are initialized: Get returns
			// while the fetch is still in flight, and a landing that beat the
			// hook would leave the memoized list stale for the session.
			catwalkCatalog.setOnLanded(onProvidersLanded)
			hyperCatalog.setOnLanded(onProvidersLanded)

			// One error variable per goroutine. This path used to append to a
			// shared slice from both, which -race would have flagged had the
			// two fetches ever failed together.
			wg.Go(func() {
				catwalkURL := cmp.Or(os.Getenv("CATWALK_URL"), defaultCatwalkURL)
				client := catwalk.NewWithURL(catwalkURL)

				catwalkCatalog.Init(client, cachePathFor("providers"), autoupdate)

				if _, err := catwalkCatalog.Get(ctx); err != nil {
					catwalkErr = fmt.Errorf("BVR was unable to fetch an updated list of providers from %s/v2/providers. Consider setting BVR_DISABLE_PROVIDER_AUTO_UPDATE=1 to use the embedded providers bundled at the time of this BVR release. You can also update providers manually. For more info see bvr-cli update-providers --help.\n\nCause: %w", catwalkURL, err)
				}
			})

			wg.Go(func() {
				hyperCatalog.Init(realHyperClient{baseURL: hyper.BaseURL()}, cachePathFor("hyper"), autoupdate)

				if _, err := hyperCatalog.Get(ctx); err != nil {
					hyperErr = fmt.Errorf("BVR was unable to fetch updated information from Hyper: %w", err)
				}
			})

			wg.Wait()
			providerErrs = []error{catwalkErr, hyperErr}
		}

		// Taken under the same lock the rebuild uses: a refresh can land while
		// this is still assembling, since Get returns before the fetch does.
		providerMu.Lock()
		providerList = currentProviderList()
		providerErr = errors.Join(providerErrs...)
		providerMu.Unlock()
	})

	providerMu.RLock()
	defer providerMu.RUnlock()
	return providerList, providerErr
}

// assembleProviders builds the list Providers serves from the two catalog
// sources. Startup and the post-refresh rebuild share this function so the
// rebuilt list has the same contents and the same order — the model dialog
// relies on Hyper and Cline leading it.
func assembleProviders(items []catwalk.Provider, hyperProvider catwalk.Provider, hasHyper bool) []catwalk.Provider {
	list := make([]catwalk.Provider, 0, len(items)+2)
	if hasHyper {
		list = append(list, hyperProvider)
	}
	list = append(list, ClineProvider())
	return append(list, items...)
}

// currentProviderList assembles the served list from whatever the catalog
// sources hold right now.
func currentProviderList() []catwalk.Provider {
	syncerMu.RLock()
	catwalkCatalog := catwalkSyncer
	hyperCatalog := hyperSyncer
	syncerMu.RUnlock()
	items, _ := catwalkCatalog.current()
	hyperProvider, hasHyper := hyperCatalog.current()
	return assembleProviders(items, hyperProvider, hasHyper)
}

// onProvidersLanded is what a catalog syncer calls once it has swapped in a
// newer result: rebuild the memoized list, then tell the listeners.
func onProvidersLanded() {
	providerMu.Lock()
	providerList = currentProviderList()
	providerMu.Unlock()

	// Bumped after the swap, so a reader that sees version N knows the list
	// matching N is already in place.
	version := providerVersion.Add(1)

	providerHooksMu.Lock()
	hooks := slices.Clone(providerHooks)
	providerHooksMu.Unlock()
	for _, fn := range hooks {
		fn(version)
	}
}

// ProviderVersion reports how many times a background catalog refresh has
// landed a newer list in this process. Zero means Providers is still serving
// the list built at startup.
func ProviderVersion() uint64 {
	return providerVersion.Load()
}

// OnProvidersUpdated registers fn to run when a background provider or Hyper
// refresh lands a newer catalog, passing the new [ProviderVersion]. fn runs
// on the goroutine that landed the data and must not block.
//
// A landing that happened before fn was registered is replayed immediately.
// The refresh is one network round trip and can finish before the TUI is
// constructed, so without the replay the UI could be told nothing and keep
// showing the startup catalog for the whole session.
func OnProvidersUpdated(fn func(version uint64)) {
	providerHooksMu.Lock()
	providerHooks = append(providerHooks, fn)
	providerHooksMu.Unlock()

	if version := providerVersion.Load(); version > 0 {
		fn(version)
	}
}

type cache[T any] struct {
	path string
}

func newCache[T any](path string) cache[T] {
	return cache[T]{path: path}
}

func (c cache[T]) Get() (T, string, error) {
	var v T
	data, err := os.ReadFile(c.path)
	if err != nil {
		return v, "", fmt.Errorf("failed to read provider cache file: %w", err)
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return v, "", fmt.Errorf("failed to unmarshal provider data from cache: %w", err)
	}

	return v, etag.Of(data), nil
}

func (c cache[T]) Store(v T) error {
	slog.Info("Saving provider data to disk", "path", c.path)
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for provider cache: %w", err)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal provider data: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write provider data to cache: %w", err)
	}
	return nil
}
