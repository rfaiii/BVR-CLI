package node

import (
	"errors"
	"net"
	"strings"
	"testing"

	zeroconf "github.com/grandcat/zeroconf"
	"github.com/stretchr/testify/require"
)

func TestPortFromAddr(t *testing.T) {
	for _, tc := range []struct {
		addr string
		want int
	}{
		{"0.0.0.0:7799", 7799},
		{"127.0.0.1:7777", 7777},
		{":8080", 8080},
		{"[::]:9000", 9000},
	} {
		got, err := portFromAddr(tc.addr)
		require.NoError(t, err, tc.addr)
		require.Equal(t, tc.want, got)
	}

	for _, addr := range []string{"", "0.0.0.0", "0.0.0.0:0", "0.0.0.0:notaport"} {
		_, err := portFromAddr(addr)
		require.Error(t, err, addr)
	}
}

// macOS reports its hostname as "Richards-MacBook-Air.local", and zeroconf
// treats a name ending in the domain as already qualified -- without the
// trailing dot every probe and announcement fails.
func TestResolveIdentityQualifiesHostName(t *testing.T) {
	identity, err := resolveIdentity()
	if errors.Is(err, ErrNoRoutableAddress) {
		t.Skip("host has no routable IPv4 address")
	}
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(identity.fqdn, "."+serviceDomain+"."), "fqdn %q", identity.fqdn)
	require.NotContains(t, identity.instance, "."+serviceDomain)
	require.NotEmpty(t, identity.addresses)
	for _, addr := range identity.addresses {
		ip := net.ParseIP(addr)
		require.NotNil(t, ip.To4(), "advertised %q is not IPv4", addr)
		require.False(t, ip.IsLinkLocalUnicast(), "link-local %q must not be advertised", addr)
	}
}

func TestNodeFromEntryReadsTXT(t *testing.T) {
	entry := &zeroconf.ServiceEntry{
		HostName: "richards-macbook-air.local.",
		Port:     7799,
		Text:     []string{"proto=1", "id=node-7", "name=Studio", "ignored"},
		AddrIPv4: []net.IP{net.ParseIP("172.20.20.20")},
	}
	entry.Instance = "Studio"

	node := nodeFromEntry(entry)
	require.Equal(t, "1", node.Protocol)
	require.Equal(t, "node-7", node.ID)
	require.Equal(t, "Studio", node.Name)
	require.Equal(t, "172.20.20.20:7799", node.HostPort())
	require.Equal(t, "ws://172.20.20.20:7799/v1/node/ws", node.WebSocketURL())
}

func TestNodeIdentityFallbacks(t *testing.T) {
	entry := &zeroconf.ServiceEntry{HostName: "host.local.", Port: 80}
	node := nodeFromEntry(entry)
	require.Equal(t, "host.local:80", node.HostPort(), "must fall back to the SRV target when no A record arrived")
	require.Equal(t, node.HostPort(), node.ID, "an unlabelled node is identified by its address")

	v6Only := Node{Addresses: []net.IP{net.ParseIP("fe80::1")}, Port: 7799}
	require.Equal(t, "[fe80::1]:7799", v6Only.HostPort())
	require.Equal(t, "ws://[fe80::1]:7799/v1/node/ws", v6Only.WebSocketURL())

	empty := Node{}
	require.Empty(t, empty.HostPort())
	require.Empty(t, empty.WebSocketURL())
}
