package beaver

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const AngleCount = 11

type BeaverRenderer struct {
	frames [AngleCount][]string
}

// NewBeaverRenderer creates the ASCII beaver.
func NewBeaverRenderer() *BeaverRenderer {
	return &BeaverRenderer{
		frames: [AngleCount][]string{
			front,
			frontLeft,
			left,
			backLeft,
			back,
			backRight,
			right,
			frontRight,
			top,
			bottom,
			tiltUp,
		},
	}
}

// Frame returns one complete frame.
// angle is automatically wrapped to 0..10.
//
// selected=true changes the eyes from O to X.
func (b *BeaverRenderer) Frame(angle int, selected bool) string {
	angle = ((angle % AngleCount) + AngleCount) % AngleCount

	lines := make([]string, len(b.frames[angle]))
	copy(lines, b.frames[angle])

	if selected {
		for i := range lines {
			lines[i] = strings.ReplaceAll(lines[i], "O", "X")
		}
	}

	return strings.Join(lines, "\n")
}

// StartAnimation continuously cycles through the beaver angles.
//
// The selected state can be changed through selectedCh.
func (b *BeaverRenderer) StartAnimation(
	ctx context.Context,
	selectedCh <-chan bool,
) {
	ticker := time.NewTicker(140 * time.Millisecond)
	defer ticker.Stop()

	angle := 0
	selected := false

	for {
		select {
		case <-ctx.Done():
			return

		case value := <-selectedCh:
			selected = value

		case <-ticker.C:
			fmt.Print("\033[H\033[2J")
			fmt.Println(b.Frame(angle, selected))

			angle++
			if angle >= AngleCount {
				angle = 0
			}
		}
	}
}

var front = []string{
	`      /\____/\      `,
	`    /@@@@@@@@@@\    `,
	`   /@@@@@@@@@@@@\   `,
	`  |@@@ O @@@ O @@@| `,
	`  |@@@@@@@@@@@@@@@@| `,
	`  |@@@@@@\__/@@@@@@| `,
	`  |@@@@@@ || @@@@@@| `,
	`   \@@@@@@@@@@@@@@/  `,
	`    \____________/   `,
}

var frontLeft = []string{
	`      /\____/\       `,
	`    /@@@@@@@@@@\     `,
	`   /@@@@@@@@@@@@\    `,
	`  |@@ O @@@@@@@@ |  `,
	`  |@@@@@@@@@@@@@@@|  `,
	`  |@@@@@\__/@@@@@@|  `,
	`  |@@@@@@ || @@@@@|  `,
	`   \@@@@@@@@@@@@@@/   `,
	`    \____________/    `,
}

var left = []string{
	`       /\___         `,
	`     /@@@@@@\        `,
	`    /@@@@@@@@@\      `,
	`   |O@@@@@@@@@@\     `,
	`   |@@@@@@@@@@@|     `,
	`   |@@@@@\__@@@@|    `,
	`   |@@@@@@||@@@@|    `,
	`    \@@@@@@@@@@/     `,
	`     \________/      `,
}

var backLeft = []string{
	`      /\____/\       `,
	`    /@@@@@@@@@@\     `,
	`   /@@@@@@@@@@@@\    `,
	`  |@@@@@@@@@@@@@@\   `,
	`  |@@@@@@@@@@@@@@@|  `,
	`  |@@@@@@@@@@@@@@@|  `,
	`  |@@@@@@@@@@@@@@@|  `,
	`   \@@@@@@@@@@@@@@/  `,
	`    \____________/   `,
}

var back = []string{
	`      /\____/\       `,
	`    /############\   `,
	`   /##############\  `,
	`  |################| `,
	`  |################| `,
	`  |################| `,
	`  |################| `,
	`   \##############/  `,
	`    \____________/   `,
}

var backRight = []string{
	`       /\____/\      `,
	`     /@@@@@@@@@@\    `,
	`    /@@@@@@@@@@@@\   `,
	`   /@@@@@@@@@@@@@@|  `,
	`  |@@@@@@@@@@@@@@@|  `,
	`  |@@@@@@@@@@@@@@@|  `,
	`  |@@@@@@@@@@@@@@@|  `,
	`   \@@@@@@@@@@@@@@/  `,
	`    \____________/   `,
}

var right = []string{
	`         ___/\       `,
	`        /@@@@@@\     `,
	`      /@@@@@@@@@\    `,
	`     /@@@@@@@@@@O|   `,
	`    |@@@@@@@@@@@@|   `,
	`    |@@@@@\__/@@@|   `,
	`    |@@@@@@||@@@@|   `,
	`     \@@@@@@@@@@/    `,
	`      \________/     `,
}

var frontRight = []string{
	`       /\____/\      `,
	`     /@@@@@@@@@@\    `,
	`    /@@@@@@@@@@@@\   `,
	`   |@@@ O @@@ O @@@| `,
	`   |@@@@@@@@@@@@@@@@|`,
	`   |@@@@@@\__/@@@@@@|`,
	`   |@@@@@@ || @@@@@@|`,
	`    \@@@@@@@@@@@@@@/ `,
	`     \____________/  `,
}

var top = []string{
	`       __________     `,
	`     /@@@@@@@@@@@@\   `,
	`   /@@@@@@@@@@@@@@@@\ `,
	`  /@@@@@@@@@@@@@@@@@@\`,
	` |@@@@@@@@@@@@@@@@@@@@|`,
	` |@@@@@@@@@@@@@@@@@@@@|`,
	`  \@@@@@@@@@@@@@@@@@@/`,
	`   \________________/ `,
}

var bottom = []string{
	`      __    __       `,
	`    /@@@@@@@@@@\     `,
	`   /@@@@@@@@@@@@\    `,
	`  |@@@@@@@@@@@@@@|   `,
	`  |@@@@@@@@@@@@@@|   `,
	`  |@@@@@@@@@@@@@@|   `,
	`   \@@@@@  @@@@@/    `,
	`    \____\/____/     `,
}

var tiltUp = []string{
	`       /\____/\      `,
	`     /@@@@@@@@@@\    `,
	`    /@@@@ O  O @@@\  `,
	`   |@@@@@@@@@@@@@@@@|`,
	`   |@@@@@@@/\@@@@@@@|`,
	`   |@@@@@@ || @@@@@@|`,
	`    \@@@@@@@@@@@@@@/ `,
	`     \____________/  `,
}
