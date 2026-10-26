//go:build (darwin || linux || windows || freebsd || openbsd || netbsd) && !ios && !android

package clipboard

import "golang.design/x/clipboard"

func initClipboard() error {
	fallbackAvailable := initFallback()
	err := clipboard.Init()
	if err != nil && fallbackAvailable {
		// X11 cannot be opened on a Wayland session without XWayland. wl-clipboard
		// covers that, so the X11 failure is not worth reporting.
		return nil
	}
	return err
}

func writeText(text string) {
	if canWrite() {
		writeFallback(text)
		return
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
}

func read(f Format) ([]byte, error) {
	if canRead() {
		return readFallback(f)
	}
	var format clipboard.Format
	switch f {
	case FormatText:
		format = clipboard.FmtText
	case FormatImage:
		format = clipboard.FmtImage
	default:
		return nil, ErrEmpty
	}
	data := clipboard.Read(format)
	if data == nil {
		return nil, ErrEmpty
	}
	return data, nil
}
