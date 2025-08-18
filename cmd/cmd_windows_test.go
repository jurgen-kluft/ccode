//go:build windows

package cmd_test

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/jurgen-kluft/ccode/cmd"
)

func TestCmdOK(t *testing.T) {
	now := time.Now().Unix()

	p := cmd.NewCmd("echo", "foo")
	gotStatus := <-p.Start()
	expectStatus := cmd.Status{
		Cmd:      "echo",
		PID:      gotStatus.PID, // nondeterministic
		Complete: true,
		Exit:     0,
		Error:    nil,
		Runtime:  gotStatus.Runtime, // nondeterministic
		Stdout:   []string{"foo"},
		Stderr:   []string{},
	}
	if gotStatus.StartTs < now {
		t.Error("StartTs < now")
	}
	if gotStatus.StopTs < gotStatus.StartTs {
		t.Error("StopTs < StartTs")
	}
	gotStatus.StartTs = 0
	gotStatus.StopTs = 0
	if diffs := deep.Equal(gotStatus, expectStatus); diffs != nil {
		t.Error(diffs)
	}
	if gotStatus.PID < 0 {
		t.Errorf("got PID %d, expected non-zero", gotStatus.PID)
	}
	if gotStatus.Runtime < 0 {
		t.Errorf("got runtime %f, expected non-zero", gotStatus.Runtime)
	}
}

func TestCmdClone(t *testing.T) {
	opt := cmd.Options{
		Buffered: true,
	}
	c1 := cmd.NewCmdOptions(opt, "ls")
	c1.Dir = "/tmp/"
	c1.Env = []string{"YES=please"}
	c2 := c1.Clone()

	if c1.Name != c2.Name {
		t.Errorf("got Name %s, expecting %s", c2.Name, c1.Name)
	}
	if c1.Dir != c2.Dir {
		t.Errorf("got Dir %s, expecting %s", c2.Dir, c1.Dir)
	}
	if diffs := deep.Equal(c1.Env, c2.Env); diffs != nil {
		t.Error(diffs)
	}
}

func TestCmdNonzeroExit(t *testing.T) {
	p := cmd.NewCmd("false")
	gotStatus := <-p.Start()
	expectStatus := cmd.Status{
		Cmd:      "false",
		PID:      gotStatus.PID, // nondeterministic
		Complete: true,
		Exit:     1,
		Error:    nil,
		Runtime:  gotStatus.Runtime, // nondeterministic
		Stdout:   []string{},
		Stderr:   []string{},
	}
	gotStatus.StartTs = 0
	gotStatus.StopTs = 0
	if diffs := deep.Equal(gotStatus, expectStatus); diffs != nil {
		t.Error(diffs)
	}
	if gotStatus.PID < 0 {
		t.Errorf("got PID %d, expected non-zero", gotStatus.PID)
	}
	if gotStatus.Runtime < 0 {
		t.Errorf("got runtime %f, expected non-zero", gotStatus.Runtime)
	}
}

func TestCmdOutput(t *testing.T) {
	t.Skip("FIXME")

	tmpfile, err := ioutil.TempFile("", "cmd.TestCmdOutput")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	t.Logf("temp file: %s", tmpfile.Name())
	os.Remove(tmpfile.Name())

	p := cmd.NewCmd(path.Join(".", "test", "touch-file-count"), tmpfile.Name())

	p.Start()

	touchFile := func(file string) {
		if err := exec.Command("touch", file).Run(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(600 * time.Millisecond)
	}
	var s cmd.Status
	var stdout []string

	touchFile(tmpfile.Name())
	s = p.Status()
	stdout = []string{"1"}
	if diffs := deep.Equal(s.Stdout, stdout); diffs != nil {
		t.Log(s.Stdout)
		t.Error(diffs)
	}

	touchFile(tmpfile.Name())
	s = p.Status()
	stdout = []string{"1", "2"}
	if diffs := deep.Equal(s.Stdout, stdout); diffs != nil {
		t.Log(s.Stdout)
		t.Error(diffs)
	}

	// No more output yet
	s = p.Status()
	stdout = []string{"1", "2"}
	if diffs := deep.Equal(s.Stdout, stdout); diffs != nil {
		t.Log(s.Stdout)
		t.Error(diffs)
	}

	// +2 lines
	touchFile(tmpfile.Name())
	touchFile(tmpfile.Name())
	s = p.Status()
	stdout = []string{"1", "2", "3", "4"}
	if diffs := deep.Equal(s.Stdout, stdout); diffs != nil {
		t.Log(s.Stdout)
		t.Error(diffs)
	}

	// Kill the process
	if err := p.Stop(); err != nil {
		t.Error(err)
	}
}

func TestCmdNotFound(t *testing.T) {
	t.Skip("FIXME")

	p := cmd.NewCmd("cmd-does-not-exist")
	gotStatus := <-p.Start()
	gotStatus.StartTs = 0
	gotStatus.StopTs = 0
	expectStatus := cmd.Status{
		Cmd:      "cmd-does-not-exist",
		PID:      0,
		Complete: false,
		Exit:     -1,
		Error:    errors.New(`exec: "cmd-does-not-exist": executable file not found in $PATH`),
		Runtime:  0,
		Stdout:   nil,
		Stderr:   nil,
	}
	if diffs := deep.Equal(gotStatus, expectStatus); diffs != nil {
		t.Logf("%+v", gotStatus)
		t.Error(diffs)
	}
}

func TestDone(t *testing.T) {
	t.Skip("FIXME")

	// Count to 3 sleeping 1s between counts
	p := cmd.NewCmd(path.Join(".", "test", "count-and-sleep"), "3", "1")
	statusChan := p.Start()

	// For 2s while cmd is running, Done() chan should block, which means
	// it's still running
	runningTimer := time.After(2 * time.Second)
TIMER:
	for {
		select {
		case <-runningTimer:
			break TIMER
		default:
		}
		select {
		case <-p.Done():
			t.Fatal("Done chan is closed before runningTime finished")
		default:
			// Done chan blocked, cmd is still running
		}
		time.Sleep(400 * time.Millisecond)
	}

	// Wait for cmd to complete
	var s1 cmd.Status
	select {
	case s1 = <-statusChan:
		t.Logf("got status: %+v", s1)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for cmd to complete")
	}

	// After cmd completes, Done chan should be closed and not block
	select {
	case <-p.Done():
	default:
		t.Fatal("Done chan did not block after cmd completed")
	}

	// After command completes, we should be able to get exact same
	// Status that's returned on the Start() chan
	s2 := p.Status()
	if diff := deep.Equal(s1, s2); diff != nil {
		t.Error(diff)
	}
}

func TestCmdEnvOK(t *testing.T) {
	t.Skip("FIXME")

	now := time.Now().Unix()

	p := cmd.NewCmd("env")
	p.Env = []string{"FOO=foo"}
	gotStatus := <-p.Start()
	expectStatus := cmd.Status{
		Cmd:      "env",
		PID:      gotStatus.PID, // nondeterministic
		Complete: true,
		Exit:     0,
		Error:    nil,
		Runtime:  gotStatus.Runtime, // nondeterministic
		Stdout:   []string{"FOO=foo"},
		Stderr:   []string{},
	}
	if gotStatus.StartTs < now {
		t.Error("StartTs < now")
	}
	if gotStatus.StopTs < gotStatus.StartTs {
		t.Error("StopTs < StartTs")
	}
	gotStatus.StartTs = 0
	gotStatus.StopTs = 0
	if diffs := deep.Equal(gotStatus, expectStatus); diffs != nil {
		t.Error(diffs)
	}
	if gotStatus.PID < 0 {
		t.Errorf("got PID %d, expected non-zero", gotStatus.PID)
	}
	if gotStatus.Runtime < 0 {
		t.Errorf("got runtime %f, expected non-zero", gotStatus.Runtime)
	}
}

func TestCmdNoOutput(t *testing.T) {
	// Set both output options to false to discard all output
	p := cmd.NewCmdOptions(
		cmd.Options{
			Buffered:  false,
			Streaming: false,
		},
		"echo", "hell-world")
	s := <-p.Start()
	if s.Exit != 0 {
		t.Errorf("got exit %d, expected 0", s.Exit)
	}
	if len(s.Stdout) != 0 {
		t.Errorf("got stdout, expected no output: %v", s.Stdout)
	}
	if len(s.Stderr) != 0 {
		t.Errorf("got stderr, expected no output: %v", s.Stderr)
	}
}
