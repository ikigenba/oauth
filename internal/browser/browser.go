// Package browser opens authorization URLs in the user's browser.
package browser

import "os/exec"

// Launcher opens a URL in the user's browser.
type Launcher interface {
	Open(url string) error
}

type command interface {
	Start() error
}

type commandFactory func(name string, args ...string) command

type commandLauncher struct {
	name       string
	newCommand commandFactory
}

func newCommandLauncher(name string) *commandLauncher {
	return &commandLauncher{
		name: name,
		newCommand: func(name string, args ...string) command {
			return exec.Command(name, args...)
		},
	}
}

func (l *commandLauncher) Open(url string) error {
	return l.newCommand(l.name, url).Start()
}
