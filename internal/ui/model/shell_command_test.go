package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The file browser hands us arbitrary paths and the bang string is parsed by a
// real shell, so an unquoted path is both a bug (spaces split the argument) and
// an injection sink (the selection is attacker-choosable content on disk).
func TestShellQuoteKeepsPathsAsOneWord(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ in, want string }{
		{"/tmp/a.png", "'/tmp/a.png'"},
		{"/tmp/my photos/cat.png", "'/tmp/my photos/cat.png'"},
		{`/tmp/it's.png`, `'/tmp/it'\''s.png'`},
		{"/tmp/$(touch pwned).png", "'/tmp/$(touch pwned).png'"},
		{"/tmp/a;rm -rf /.png", "'/tmp/a;rm -rf /.png'"},
	} {
		require.Equal(t, tc.want, shellQuote(tc.in), "quoting %q", tc.in)
	}
}
