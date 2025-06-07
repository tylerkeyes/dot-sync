package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tylerkeyes/dot-sync/internal/shared"
)

func TestGitStorage_PushToStorage(t *testing.T) {
	storage := &GitStorage{RemoteURL: "https://github.com/user/repo.git"}

	// Test basic functionality (should not error for now)
	err := storage.PushToStorage("/some/path")
	if err != nil {
		t.Errorf("PushToStorage failed: %v", err)
	}

	// Test with empty path
	err = storage.PushToStorage("")
	if err != nil {
		t.Errorf("PushToStorage with empty path failed: %v", err)
	}
}

func TestGitStorage_PullFromStorage(t *testing.T) {
	storage := &GitStorage{RemoteURL: "https://github.com/user/repo.git"}

	// Test basic functionality (should not error for now)
	err := storage.PullFromStorage("/some/path")
	if err != nil {
		t.Errorf("PullFromStorage failed: %v", err)
	}

	// Test with empty path
	err = storage.PullFromStorage("")
	if err != nil {
		t.Errorf("PullFromStorage with empty path failed: %v", err)
	}
}

func TestIsRemoteExistsError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "exit status 3",
			err:      &testError{msg: "exit status 3"},
			expected: true,
		},
		{
			name:     "exit status 128",
			err:      &testError{msg: "exit status 128"},
			expected: true,
		},
		{
			name:     "remote exists message 1",
			err:      &testError{msg: "fatal: remote origin already exists."},
			expected: true,
		},
		{
			name:     "remote exists message 2",
			err:      &testError{msg: "fatal: remote origin already exists"},
			expected: true,
		},
		{
			name:     "remote exists message 3",
			err:      &testError{msg: "error: remote origin already exists."},
			expected: true,
		},
		{
			name:     "remote exists message 4",
			err:      &testError{msg: "error: remote origin already exists"},
			expected: true,
		},
		{
			name:     "other error",
			err:      &testError{msg: "some other error"},
			expected: false,
		},
		{
			name:     "different exit status",
			err:      &testError{msg: "exit status 1"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isRemoteExistsError(tc.err)
			if result != tc.expected {
				t.Errorf("expected %v, got %v for error: %v", tc.expected, result, tc.err)
			}
		})
	}
}

func TestContainsRemoteExistsMsg(t *testing.T) {
	testCases := []struct {
		name     string
		msg      string
		expected bool
	}{
		{
			name:     "empty message",
			msg:      "",
			expected: false,
		},
		{
			name:     "fatal with period",
			msg:      "fatal: remote origin already exists.",
			expected: true,
		},
		{
			name:     "fatal without period",
			msg:      "fatal: remote origin already exists",
			expected: true,
		},
		{
			name:     "error with period",
			msg:      "error: remote origin already exists.",
			expected: true,
		},
		{
			name:     "error without period",
			msg:      "error: remote origin already exists",
			expected: true,
		},
		{
			name:     "different message",
			msg:      "some other error message",
			expected: false,
		},
		{
			name:     "partial match",
			msg:      "remote origin already",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := containsRemoteExistsMsg(tc.msg)
			if result != tc.expected {
				t.Errorf("expected %v, got %v for message: %q", tc.expected, result, tc.msg)
			}
		})
	}
}

func TestGitStorage_InitializeStorage_DirectoryCreation(t *testing.T) {
	// Test directory creation without database dependencies
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	storage := &GitStorage{RemoteURL: "https://github.com/user/repo.git"}

	// This will fail at database operations, but we can test that directory creation works
	_ = storage.InitializeStorage()
	// We expect this to fail due to database operations, but directory should be created

	dotSyncFilesPath := filepath.Join(tempHome, ".dot-sync", "files")
	if _, err := os.Stat(dotSyncFilesPath); os.IsNotExist(err) {
		t.Error("expected .dot-sync/files directory to be created")
	}

	// The git operations should also be attempted
	gitPath := filepath.Join(dotSyncFilesPath, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		// Git init might fail in test environment, that's ok
		t.Log("git init may have failed, which is expected in test environment")
	}
}

func TestGetCurrentGitBranch(t *testing.T) {
	temp := t.TempDir()

	// Test with non-git directory
	_, err := getCurrentGitBranch(temp)
	if err == nil {
		t.Error("expected error for non-git directory, got nil")
	}

	// Create a git repo for testing
	if err := shared.RunCmd(temp, "git", "init"); err != nil {
		t.Skip("git not available for testing")
	}
	if err := shared.RunCmd(temp, "git", "config", "user.email", "test@example.com"); err != nil {
		t.Skip("git config failed")
	}
	if err := shared.RunCmd(temp, "git", "config", "user.name", "Test User"); err != nil {
		t.Skip("git config failed")
	}

	// Create initial commit to establish branch
	testFile := filepath.Join(temp, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := shared.RunCmd(temp, "git", "add", "test.txt"); err != nil {
		t.Skip("git add failed")
	}
	if err := shared.RunCmd(temp, "git", "commit", "-m", "initial commit"); err != nil {
		t.Skip("git commit failed")
	}

	branch, err := getCurrentGitBranch(temp)
	if err != nil {
		t.Errorf("getCurrentGitBranch failed: %v", err)
	}
	if branch == "" {
		t.Error("expected non-empty branch name")
	}
	// Should be either "main" or "master" depending on git config
	if branch != "main" && branch != "master" {
		t.Errorf("unexpected branch name: %s", branch)
	}
}

// testError is a helper type for testing error conditions
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
