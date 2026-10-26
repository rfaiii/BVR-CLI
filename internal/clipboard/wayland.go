package clipboard

// The Wayland clipboard is reached through the wl-clipboard CLI rather than the
// X11/cgo backend in golang.design/x/clipboard, which cannot talk to a Wayland
// compositor at all without XWayland. On Omarchy that means Hyprland with no
// XWayland running, where X11 writes currently go nowhere.
//
// The decision logic lives here, without a build tag, so it can be tested on any
// host; only the exec side is Linux-only.

// backend records which wl-clipboard binaries were found.
type backend struct {
	copyBin  string
	pasteBin string
}

func (b backend) canWrite() bool { return b.copyBin != "" }
func (b backend) canRead() bool  { return b.pasteBin != "" }

// detect chooses the wl-clipboard binaries for a session. Wayland has to be
// signalled and at least one binary present: under X11 the native backend is
// strictly better than shelling out.
func detect(waylandDisplay string, lookPath func(string) (string, error)) backend {
	if waylandDisplay == "" {
		return backend{}
	}
	var found backend
	if path, err := lookPath("wl-copy"); err == nil {
		found.copyBin = path
	}
	if path, err := lookPath("wl-paste"); err == nil {
		found.pasteBin = path
	}
	if found.copyBin == "" && found.pasteBin == "" {
		return backend{}
	}
	return found
}

// pasteArgs holds the wl-paste flags. `-n` matters: without it a trailing
// newline is appended, which would corrupt a pasted path. An image is asked for
// by MIME type, so a clipboard holding a different encoding comes back empty
// rather than as unrelated bytes.
func pasteArgs(f Format) ([]string, bool) {
	switch f {
	case FormatText:
		return []string{"-n"}, true
	case FormatImage:
		return []string{"--type", "image/png"}, true
	default:
		return nil, false
	}
}
