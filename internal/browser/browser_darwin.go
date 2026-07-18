//go:build darwin

package browser

// New returns the platform browser launcher.
func New() Launcher {
	return newCommandLauncher("open")
}
