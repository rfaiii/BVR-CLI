//go:build !darwin

package notification

import (
	_ "embed"
)

//go:embed bvr-icon-solo.png
var Icon []byte
