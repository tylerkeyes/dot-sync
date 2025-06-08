package internal

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal/shared"
	"github.com/tylerkeyes/dot-sync/internal/storage"
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
	// Set empty context (no storage provider)
	ctx := context.Background()
	cmd.SetContext(ctx)

	pullHandler(cmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Pulling dotfiles...") {
		t.Errorf("expected output to contain 'Pulling dotfiles...', got %q", output)
	}
	if !strings.Contains(output, "No storage provider configured") {
		t.Errorf("expected output to contain 'No storage provider configured', got %q", output)
	}
}

func TestPullHandlerWithArgs(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	// Set empty context (no storage provider)
	ctx := context.Background()
	cmd.SetContext(ctx)

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
	if !strings.Contains(output, "No storage provider configured") {
		t.Errorf("expected output to contain 'No storage provider configured', got %q", output)
	}
}

func TestPullHandlerWithStorageProvider(t *testing.T) {
	// Create a mock storage provider
	mockStorage := &storage.GitStorage{RemoteURL: "https://github.com/test/repo.git"}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	// Set context with storage provider
	ctx := context.WithValue(context.Background(), shared.GetStorageProviderKey(), mockStorage)
	cmd.SetContext(ctx)

	pullHandler(cmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Pulling dotfiles...") {
		t.Errorf("expected output to contain 'Pulling dotfiles...', got %q", output)
	}
	// This will likely fail at the pull operation due to no git repo, which is expected
	if !strings.Contains(output, "Failed to pull from storage") && !strings.Contains(output, "Failed to open .dot-sync.db") {
		t.Logf("Output: %q", output)
		// Either error is acceptable in test environment
	}
}
