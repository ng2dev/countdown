package main

import (
	"context"
	"os"
	"os/exec"
	"time"

	tmtest "github.com/iov-one/weave/tmtest"
)

// RunCountdownd is like RunTendermint, just executes the countdown executable, assuming a prepared home directory
// This package will be removed with weave 0.22.0 since it is a copy-paste from the repo
func RunCountdownd(ctx context.Context, t tmtest.TestReporter, home string) (cleanup func()) {
	t.Helper()

	appPath, err := exec.LookPath("countdown")
	if err != nil {
		if os.Getenv("FORCE_TM_TEST") != "1" {
			t.Skip("countdown binary not found. Set FORCE_TM_TEST=1 to fail this test.")
		} else {
			t.Fatalf("%s binary not found. Do not set FORCE_TM_TEST=1 to skip this test.", "countdown")
		}
	}

	cmd := exec.CommandContext(ctx, appPath, "-home", home, "start")
	// log tendermint output for verbose debugging....
	if os.Getenv("TM_DEBUG") != "" {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("%s process failed: %s", "countdown", err)
	}

	// Give tendermint time to setup.
	time.Sleep(2 * time.Second)
	t.Logf("Running %s pid=%d", appPath, cmd.Process.Pid)

	// Return a cleanup function, that will wait for app to stop.
	// We also auto-kill when the context is Done
	cleanup = func() {
		t.Logf("%s cleanup called", "countdown")
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
	go func() {
		<-ctx.Done()
		cleanup()
	}()
	return cleanup
}
