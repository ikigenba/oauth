//go:build darwin

package browser

import (
	"reflect"
	"testing"
)

type recordingCommand struct {
	started bool
}

func (c *recordingCommand) Start() error {
	c.started = true
	return nil
}

// R-1INO-S925
func TestDarwinOpenInvokesOpenWithUnmodifiedURLAsSoleArgument(t *testing.T) {
	launcher := New().(*commandLauncher)
	wantURL := "https://issuer.example/authorize?scope=openid+profile&state=a%2Fb c"

	var gotName string
	var gotArgs []string
	cmd := &recordingCommand{}
	launcher.newCommand = func(name string, args ...string) command {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return cmd
	}

	if err := launcher.Open(wantURL); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if gotName != "open" {
		t.Errorf("command name = %q, want %q", gotName, "open")
	}
	if want := []string{wantURL}; !reflect.DeepEqual(gotArgs, want) {
		t.Errorf("command arguments = %#v, want %#v", gotArgs, want)
	}
	if !cmd.started {
		t.Error("command was not started")
	}
}
