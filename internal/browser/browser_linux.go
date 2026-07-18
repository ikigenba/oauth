//go:build linux

package browser

// New returns the platform browser launcher.
func New() Launcher {
	return newCommandLauncher("xdg-open")
}
