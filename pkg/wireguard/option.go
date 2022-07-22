package wireguard

import "io"

type WGOption func(*WGQuickManager)

func WithOutput(w io.Writer) WGOption {
	return func(wm *WGQuickManager) {
		wm.logOutput = w
	}
}
