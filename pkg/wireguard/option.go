package wireguard

import "io"

type WGOption func(*WGQuickManager)

func WithOutput(w io.Writer) WGOption {
	return func(wm *WGQuickManager) {
		wm.logOutput = w
	}
}

func WithPost(postUp, postDown string) WGOption {
	return func(wm *WGQuickManager) {
		wm.postup = &postUp
		wm.postdown = &postDown
	}
}
