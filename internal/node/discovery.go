package node

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	zeroconf "github.com/grandcat/zeroconf"
)

const (
	// ServiceType is the DNS-SD service type a NODE agent advertises. The
	// iOS companion declares the same value in NSBonjourServices, so the two
	// have to change together.
	ServiceType   = "_bvr"
	serviceDomain = "local"

	// ProtocolVersion is the NODE frame protocol this build speaks, published
	// in the TXT record so a client can skip a host it cannot talk to.
	ProtocolVersion = "1"

	txtProtocol = "proto"
	txtID       = "id"
	txtName     = "name"

	// BrowseWindow is how long a scan listens before reporting what it saw.
	// mDNS answers arrive out of order (PTR, then SRV, then A), so anything
	// much shorter than this drops live hosts.
	BrowseWindow = 3 * time.Second
)

// ErrNoRoutableAddress means this host cannot advertise a NODE agent that
// anything on the network could actually reach.
var ErrNoRoutableAddress = errors.New("no routable IPv4 address found")

// Node is a BVR host found on the local network.
type Node struct {
	Instance  string
	HostName  string
	Addresses []net.IP
	Port      int
	Protocol  string
	ID        string
	Name      string
}

// HostPort returns the dial target, preferring IPv4 because that is what a
// phone on a home network can reliably reach.
func (n Node) HostPort() string {
	for _, ip := range n.Addresses {
		if v4 := ip.To4(); v4 != nil {
			return net.JoinHostPort(v4.String(), strconv.Itoa(n.Port))
		}
	}
	if len(n.Addresses) > 0 {
		return net.JoinHostPort(n.Addresses[0].String(), strconv.Itoa(n.Port))
	}
	if n.HostName != "" {
		return net.JoinHostPort(strings.TrimSuffix(n.HostName, "."), strconv.Itoa(n.Port))
	}
	return ""
}

// WebSocketURL is the endpoint the companion app dials. The NODE agent
// upgrades it after validating the bearer token.
func (n Node) WebSocketURL() string {
	hostPort := n.HostPort()
	if hostPort == "" {
		return ""
	}
	// HostPort is built with net.JoinHostPort, so an IPv6 literal already
	// carries the brackets a ws:// URL needs.
	return "ws://" + hostPort + websocketPath
}

func (n Node) String() string {
	return fmt.Sprintf("%s (%s)", n.DisplayName(), n.HostPort())
}

// DisplayName is the published label for a node, falling back to whatever
// identifier the advertisement carried.
func (n Node) DisplayName() string {
	return firstNonEmpty(n.Name, n.Instance, n.HostPort())
}

// AdvertiseOptions configures mDNS publication for a running NODE agent.
type AdvertiseOptions struct {
	// Addr is the agent's listen address, used only for its port.
	Addr string
	// Instance and Name default to the machine hostname.
	Instance string
	Name     string
	// ID is a stable identifier clients use to recognise a returning host.
	// It defaults to the hostname and must not contain secrets.
	ID string
}

// Advertisement is a live mDNS publication. Closing it withdraws the record.
type Advertisement struct {
	// Instance is the name the node answers queries under.
	Instance string

	server *zeroconf.Server
}

// Close stops answering queries for this node.
func (a *Advertisement) Close() error {
	a.server.Shutdown()
	return nil
}

// Advertise publishes a NODE agent over mDNS until the returned advertisement
// is closed. The TXT record carries the protocol version and node ID only; the
// bearer token is deliberately never published.
func Advertise(opts AdvertiseOptions) (*Advertisement, error) {
	port, err := portFromAddr(opts.Addr)
	if err != nil {
		return nil, err
	}
	identity, err := resolveIdentity()
	if err != nil {
		return nil, err
	}
	instance := firstNonEmpty(opts.Instance, identity.instance)
	name := firstNonEmpty(opts.Name, identity.instance)
	id := firstNonEmpty(opts.ID, identity.instance)

	// zeroconf.Register derives the host name from os.Hostname(), which on
	// macOS already ends in ".local"; the library then decides the name is
	// qualified, leaves the trailing dot off, and every probe and announcement
	// fails with "domain must be fully qualified". RegisterProxy takes the
	// name and addresses we already resolved, so advertise through that.
	server, err := zeroconf.RegisterProxy(instance, ServiceType, serviceDomain, port, identity.fqdn, identity.addresses, []string{
		txtProtocol + "=" + ProtocolVersion,
		txtID + "=" + id,
		txtName + "=" + name,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("advertising NODE %s over mDNS: %w", instance, err)
	}
	return &Advertisement{Instance: instance, server: server}, nil
}

type localIdentity struct {
	instance  string
	fqdn      string
	addresses []string
}

// resolveIdentity returns the names and routable addresses to advertise. The
// addresses matter: a companion dials whatever the A record says, so link-local
// and loopback addresses are left out.
func resolveIdentity() (localIdentity, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return localIdentity{}, fmt.Errorf("resolving hostname for mDNS: %w", err)
	}
	base := strings.TrimSuffix(hostname, "."+serviceDomain)
	identity := localIdentity{instance: base, fqdn: base + "." + serviceDomain + "."}

	ifaces, err := net.Interfaces()
	if err != nil {
		return identity, nil
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil || ip.To4() == nil || !ip.IsGlobalUnicast() {
				continue
			}
			identity.addresses = append(identity.addresses, ip.String())
		}
	}
	if len(identity.addresses) == 0 {
		return localIdentity{}, fmt.Errorf("%w: connect this host to the network the companion is on", ErrNoRoutableAddress)
	}
	return identity, nil
}

// Browse listens for advertised NODE agents for the lifetime of ctx and
// returns what it found once ctx is done.
func Browse(ctx context.Context) ([]Node, error) {
	resolver, err := zeroconf.NewResolver(zeroconf.SelectIPTraffic(zeroconf.IPv4))
	if err != nil {
		return nil, fmt.Errorf("starting mDNS resolver: %w", err)
	}
	entries := make(chan *zeroconf.ServiceEntry)
	// Browse blocks until ctx is cancelled, closing entries behind it.
	browseErr := make(chan error, 1)
	go func() { browseErr <- resolver.Browse(ctx, ServiceType, serviceDomain, entries) }()

	found := make(map[string]Node)
	for entry := range entries {
		node := nodeFromEntry(entry)
		if existing, ok := found[node.ID]; ok && len(existing.Addresses) > 0 && len(node.Addresses) == 0 {
			continue
		}
		found[node.ID] = node
	}
	// A resolver that never started leaves the channel closed with nothing in
	// it, which would otherwise look like an empty network.
	if len(found) == 0 {
		select {
		case err := <-browseErr:
			if err != nil {
				return nil, fmt.Errorf("browsing for NODEs: %w", err)
			}
		default:
		}
	}

	nodes := make([]Node, 0, len(found))
	for _, node := range found {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].HostPort() < nodes[j].HostPort() })
	return nodes, nil
}

// Scan lists NODE agents on the local network, waiting window for late
// records. A zero window uses BrowseWindow.
func Scan(ctx context.Context, window time.Duration) ([]Node, error) {
	if window <= 0 {
		window = BrowseWindow
	}
	ctx, cancel := context.WithTimeout(ctx, window)
	defer cancel()
	return Browse(ctx)
}

func nodeFromEntry(entry *zeroconf.ServiceEntry) Node {
	node := Node{
		Instance: entry.Instance,
		HostName: entry.HostName,
		Port:     entry.Port,
	}
	for _, ip := range entry.AddrIPv4 {
		node.Addresses = append(node.Addresses, ip)
	}
	for _, ip := range entry.AddrIPv6 {
		node.Addresses = append(node.Addresses, ip)
	}
	for _, record := range entry.Text {
		key, value, found := strings.Cut(record, "=")
		if !found {
			continue
		}
		switch key {
		case txtProtocol:
			node.Protocol = value
		case txtID:
			node.ID = value
		case txtName:
			node.Name = value
		}
	}
	if node.ID == "" {
		node.ID = firstNonEmpty(node.HostPort(), node.Instance)
	}
	return node
}

func portFromAddr(addr string) (int, error) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		// A bare ":7777" or "7777" is a valid listen address with no host.
		if strings.HasPrefix(addr, ":") {
			return strconv.Atoi(strings.TrimPrefix(addr, ":"))
		}
		return 0, fmt.Errorf("parsing listen address %q: %w", addr, err)
	}
	num, err := strconv.Atoi(port)
	if err != nil || num <= 0 || num > 65535 {
		return 0, fmt.Errorf("listen address %q has no usable port", addr)
	}
	return num, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
