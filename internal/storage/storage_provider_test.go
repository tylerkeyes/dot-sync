package storage

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewStorageProviderCmd(t *testing.T) {
	cmd := NewStorageProviderCmd()
	if cmd == nil {
		t.Fatal("NewStorageProviderCmd returned nil")
	}
	if !strings.Contains(cmd.Use, "storage") {
		t.Errorf("expected Use to contain 'storage', got %q", cmd.Use)
	}
	if !strings.Contains(cmd.Short, "storage backends") {
		t.Errorf("expected Short to contain 'storage backends', got %q", cmd.Short)
	}

	// Should have subcommands
	if len(cmd.Commands()) == 0 {
		t.Error("expected storage command to have subcommands")
	}

	// Check for init subcommand
	var initCmd *cobra.Command
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == "init" {
			initCmd = subcmd
			break
		}
	}
	if initCmd == nil {
		t.Error("expected init subcommand to exist")
	}
}

func TestNewInitCmd(t *testing.T) {
	cmd := newInitCmd()
	if cmd == nil {
		t.Fatal("newInitCmd returned nil")
	}
	if !strings.Contains(cmd.Use, "init") {
		t.Errorf("expected Use to contain 'init', got %q", cmd.Use)
	}
	if !strings.Contains(cmd.Short, "Initialize backend") {
		t.Errorf("expected Short to contain 'Initialize backend', got %q", cmd.Short)
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}

	// Check for flags
	providerFlag := cmd.Flags().Lookup("provider")
	if providerFlag == nil {
		t.Error("expected --provider flag to exist")
	}
	if providerFlag.DefValue != "git" {
		t.Errorf("expected --provider default to be 'git', got %q", providerFlag.DefValue)
	}

	remoteFlag := cmd.Flags().Lookup("remote-url")
	if remoteFlag == nil {
		t.Error("expected --remote-url flag to exist")
	}
}

func TestInitCmdExecution(t *testing.T) {
	cmd := newInitCmd()

	// Test without remote URL (should fail)
	cmd.SetArgs([]string{"--provider", "git"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --remote-url is missing, got nil")
	}
	if !strings.Contains(err.Error(), "remote-url is required") {
		t.Errorf("expected error message about remote-url, got %q", err.Error())
	}

	// Test with unsupported provider
	cmd = newInitCmd()
	cmd.SetArgs([]string{"--provider", "unsupported", "--remote-url", "https://example.com"})
	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for unsupported provider, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported storage provider") {
		t.Errorf("expected error message about unsupported provider, got %q", err.Error())
	}

	// Test with git provider and remote URL would require database setup
	// This is more of an integration test, so we'll skip detailed execution testing
}
