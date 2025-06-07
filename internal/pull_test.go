package internal

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewPullCmd(t *testing.T) {
	cmd := NewPullCmd()
	if cmd == nil {
		t.Fatal("NewPullCmd returned nil")
	}
	if !strings.Contains(cmd.Use, "pull") {
		t.Errorf("expected Use to contain 'pull', got %q", cmd.Use)
	}
	if cmd.Run == nil {
		t.Error("expected Run to be set")
	}
	if !strings.Contains(cmd.Short, "Pull dotfiles") {
		t.Errorf("expected Short to contain 'Pull dotfiles', got %q", cmd.Short)
	}
}

func TestPullHandler(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	pullHandler(cmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Pulling dotfiles...") {
		t.Errorf("expected output to contain 'Pulling dotfiles...', got %q", output)
	}
}

func TestPullHandlerWithArgs(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	pullHandler(cmd, []string{"arg1", "arg2"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Args should be ignored, same output as no args
	if !strings.Contains(output, "Pulling dotfiles...") {
		t.Errorf("expected output to contain 'Pulling dotfiles...', got %q", output)
	}
}
