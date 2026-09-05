package node

import (
	"bytes"
	"os/exec"
)

type Runtime struct {
	Name      string
	Available bool
	Version   string
}

func Detect() []Runtime {
	result := make([]Runtime, 0, 7)
	for _, name := range []string{"node", "bun", "deno", "npx", "npm", "pnpm", "yarn"} {
		if _, err := exec.LookPath(name); err != nil {
			continue
		}
		runtime := Runtime{Name: name, Available: true}
		if output, err := exec.Command(name, "--version").Output(); err == nil {
			runtime.Version = string(bytes.TrimSpace(output))
		}
		result = append(result, runtime)
	}
	return result
}
