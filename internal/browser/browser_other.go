//go:build !linux && !darwin

package browser

import "errors"

type unsupportedLauncher struct{}

func (unsupportedLauncher) Open(string) error {
	return errors.New("no supported browser launcher")
}

// New returns a launcher that reports the platform is unsupported.
func New() Launcher {
	return unsupportedLauncher{}
}
