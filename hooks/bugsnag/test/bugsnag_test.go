package bugsnag

import (
	"errors"
	"os"
	"testing"
	"time"

	bugsnagHook "github.com/hatchify/output/hooks/bugsnag"

	"github.com/hatchify/output"
)

func TestBugsnagHook(t *testing.T) {
	opts := &bugsnagHook.HookOptions{
		Env:        "test",
		AppVersion: "magic_horse",
	}
	out := output.NewOutputter(os.Stderr, new(output.TextFormatter), bugsnagHook.NewHook(opts))
	out.Info("test has started")
	out.Error("1) some fake error as text")
	output.Warning("2) also default outputter with enabled env")
	out.WithError(errors.New("some fake error")).WithFields(output.Fields{
		"@user.name": "Max",
	}).Error("3) with fields and error, also meta")
	time.Sleep(time.Second)
	out.Debug("test done")
}
