package node

import (
	"fmt"
	"strings"
)

// StartupInfo is the banner printed by `node serve` as soon as the
// listener is configured, including when --advertise is not set.
type StartupInfo struct {
	Addr         string
	Exec         ExecPolicy
	Projects     []ProjectInfo
	WatchWarning string
}

// Banner is the startup text. The first line is always the listen
// address so a localhost agent is not silent.
func (s StartupInfo) Banner() string {
	var b strings.Builder
	fmt.Fprintf(&b, "NODE agent listening on %s\n", s.Addr)
	if s.Exec.Unrestricted {
		b.WriteString("Execute: unrestricted (any command). Omit --unrestricted-exec to use the allow-list.\n")
	} else {
		fmt.Fprintf(&b, "Execute allow-list: %s\n", strings.Join(s.Exec.Names(), ", "))
	}
	if len(s.Projects) == 0 {
		b.WriteString("No projects shared. Pass --project <path> so another machine can read files.\n")
	}
	for _, project := range s.Projects {
		fmt.Fprintf(&b, "Sharing project %s (%s). Peers can read and write every file under this root.\n", project.ID, project.Root)
	}
	if s.WatchWarning != "" {
		b.WriteString(s.WatchWarning + "\n")
	}
	return b.String()
}
