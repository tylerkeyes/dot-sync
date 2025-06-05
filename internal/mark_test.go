package internal

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestMarkCommandArgs(t *testing.T) {
	var receivedArgs []string
	cmd := &cobra.Command{
		Use:   "mark [files or directories]...",
		Short: "Mark a file or directory for syncing",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			receivedArgs = args
		},
	}

	inputs := []string{"file1.txt", "file2.txt", "dir1"}
	cmd.SetArgs(inputs)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(receivedArgs) != len(inputs) {
		t.Fatalf("expected %d args, got %d", len(inputs), len(receivedArgs))
	}
	for i, v := range inputs {
		if receivedArgs[i] != v {
			t.Errorf("expected arg %d to be %q, got %q", i, v, receivedArgs[i])
		}
	}
}

func TestNewMarkCmd(t *testing.T) {
	cmd := NewMarkCmd()
	if cmd == nil {
		t.Fatal("NewMarkCmd returned nil")
	}
	if !strings.Contains(cmd.Use, "mark") {
		t.Errorf("expected Use to contain 'mark', got %q", cmd.Use)
	}
	if cmd.Run == nil {
		t.Error("expected Run to be set")
	}
}

func TestArgsAsFullPaths(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get current working dir: %v", err)
	}

	relPath := "relfile.txt"
	singleName := "file.txt"
	absInput := filepath.Join(string(os.PathSeparator), "tmp", "absinput.txt")

	args := []string{singleName, relPath, absInput}
	result := argsAsFullPaths(args)

	expected := []string{
		filepath.Join(cwd, singleName),
		filepath.Join(cwd, relPath),
		absInput,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestArgsAsFullPathsEdgeCases(t *testing.T) {
	// Empty args
	if res := argsAsFullPaths([]string{}); len(res) != 0 {
		t.Errorf("expected empty result for empty input, got %v", res)
	}

	// Only absolute paths
	abs := filepath.Join(string(os.PathSeparator), "tmp", "foo")
	if res := argsAsFullPaths([]string{abs}); !reflect.DeepEqual(res, []string{abs}) {
		t.Errorf("expected %v, got %v", []string{abs}, res)
	}

	// Only single name
	cwd, _ := os.Getwd()
	if res := argsAsFullPaths([]string{"bar"}); !reflect.DeepEqual(res, []string{filepath.Join(cwd, "bar")}) {
		t.Errorf("expected %v, got %v", []string{filepath.Join(cwd, "bar")}, res)
	}
}

func TestMarkHandlerNoArgs(t *testing.T) {
	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	markHandler(cmd, []string{})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()
	if !strings.Contains(output, "No changes.") {
		t.Errorf("expected output to contain 'No changes.', got %q", output)
	}
}
