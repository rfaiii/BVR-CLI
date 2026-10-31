package node

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// DefaultAllowExec is the command allow-list used by `node serve` unless
// --unrestricted-exec is set. uname stays on the list so the iOS
// companion's existing connection test keeps working.
var DefaultAllowExec = []string{
	"uname", "node", "npm", "npx", "pnpm", "yarn", "bun", "deno",
}

// ExecPolicy decides which command names the execute endpoint may run.
// The zero value is the default allow-list. Unrestricted restores the
// previous behavior, where any command name was accepted.
type ExecPolicy struct {
	Unrestricted bool
	Extra        []string
}

// Names reports the commands a client should expect. Unrestricted mode
// reports a single "*" so capabilities stay honest.
func (p ExecPolicy) Names() []string {
	if p.Unrestricted {
		return []string{"*"}
	}
	out := append([]string{}, DefaultAllowExec...)
	seen := map[string]struct{}{}
	for _, name := range out {
		seen[commandBase(name)] = struct{}{}
	}
	for _, name := range p.Extra {
		base := commandBase(name)
		if base == "" {
			continue
		}
		if _, ok := seen[base]; ok {
			continue
		}
		seen[base] = struct{}{}
		out = append(out, base)
	}
	return out
}

// Allows reports whether command may be executed. The check uses the
// base name, so /usr/bin/uname matches uname. It is not a sandbox:
// an allowed program can still open any path on MAIN.
func (p ExecPolicy) Allows(command string) error {
	base := commandBase(command)
	if base == "" || base == "." || base == ".." {
		return fmt.Errorf("command %q is not allowed", command)
	}
	if p.Unrestricted {
		return nil
	}
	for _, name := range p.Names() {
		if name == base {
			return nil
		}
	}
	return fmt.Errorf("command %q is not on the NODE execute allow-list (pass --allow-exec %s or --unrestricted-exec)", base, base)
}

func commandBase(command string) string {
	base := filepath.Base(strings.TrimSpace(command))
	base = strings.TrimSuffix(strings.ToLower(base), ".exe")
	if strings.ContainsRune(base, 0) {
		return ""
	}
	return base
}

// tokensEqual compares two bearer tokens without leaking the length or
// the first differing byte through a short-circuit compare.
func tokensEqual(a, b string) bool {
	sumA := sha256.Sum256([]byte(a))
	sumB := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(sumA[:], sumB[:]) == 1
}

func bearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "bearer "
	if len(header) >= len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}

func authorizedBearer(expected string, r *http.Request) bool {
	if expected == "" {
		return true
	}
	return tokensEqual(expected, bearerToken(r))
}

// originAllowed accepts native clients, which send no Origin, and
// browser clients whose Origin host matches the request Host. A page
// on another origin cannot open the socket just because it guessed
// the URL.
func originAllowed(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Host, r.Host)
}

// ListenHostIsLoopback reports whether addr is a loopback listen
// address. Advertising one would publish a LAN IP for a port that is
// only open on the machine itself.
func ListenHostIsLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// LocalNodeName is the id published in /v1/node/health and in mDNS.
// macOS hostnames arrive as "name.local"; the suffix is stripped so
// the health id matches the Bonjour instance.
func LocalNodeName() string {
	host, err := os.Hostname()
	if err != nil {
		return "bvr-node"
	}
	host = strings.TrimSpace(strings.TrimSuffix(host, ".local"))
	if host == "" {
		return "bvr-node"
	}
	return host
}
