package git

import (
	"os"
	"os/exec"
	"testing"
)

// TestGetActivityInGitRepo runs GetActivity against the stacklit repository
// itself.  It skips if git is not available on PATH.
func TestGetActivityInGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}

	// Walk up from this file's package directory to find the repo root.
	// The package lives at internal/git/ so the repo root is two levels up.
	root, err := repoRoot()
	if err != nil {
		t.Skip("could not determine repo root:", err)
	}

	act, err := GetActivity(root, 90)
	if err != nil {
		t.Fatalf("GetActivity returned error: %v", err)
	}
	if act == nil {
		t.Fatal("GetActivity returned nil Activity")
	}

	t.Logf("HotFiles:    %d entries", len(act.HotFiles))
	t.Logf("RecentFiles: %d entries", len(act.RecentFiles))
	t.Logf("StableFiles: %d entries", len(act.StableFiles))

	// Invariant: no list should exceed its documented maximum.
	if len(act.HotFiles) > 20 {
		t.Errorf("HotFiles has %d entries, want ≤20", len(act.HotFiles))
	}
	if len(act.RecentFiles) > 10 {
		t.Errorf("RecentFiles has %d entries, want ≤10", len(act.RecentFiles))
	}
	if len(act.StableFiles) > 10 {
		t.Errorf("StableFiles has %d entries, want ≤10", len(act.StableFiles))
	}
}

// TestGetActivityNotGitRepo verifies that a non-git directory returns an
// empty Activity without an error.
func TestGetActivityNotGitRepo(t *testing.T) {
	dir := t.TempDir()
	act, err := GetActivity(dir, 90)
	if err != nil {
		t.Fatalf("expected no error for non-git dir, got: %v", err)
	}
	if act == nil {
		t.Fatal("expected non-nil Activity")
	}
	if len(act.HotFiles) != 0 || len(act.RecentFiles) != 0 || len(act.StableFiles) != 0 {
		t.Error("expected all lists to be empty for non-git directory")
	}
}

// TestMerkleHash verifies determinism and change-sensitivity.
func TestMerkleHash(t *testing.T) {
	files := []string{"a.go", "b.go", "c.go"}
	contents := map[string][]byte{
		"a.go": []byte("package a"),
		"b.go": []byte("package b"),
		"c.go": []byte("package c"),
	}

	h1 := ComputeMerkle(files, contents)
	h2 := ComputeMerkle(files, contents)

	if h1 == "" {
		t.Fatal("expected non-empty hash")
	}
	if h1 != h2 {
		t.Errorf("same inputs produced different hashes: %s vs %s", h1, h2)
	}

	// Mutate one file — hash must change.
	contents["b.go"] = []byte("package b_changed")
	h3 := ComputeMerkle(files, contents)
	if h3 == h1 {
		t.Error("changed file content did not change the Merkle root")
	}

	// Restore and change a path — hash must change.
	contents["b.go"] = []byte("package b")
	filesAlt := []string{"a.go", "b_renamed.go", "c.go"}
	contentsAlt := map[string][]byte{
		"a.go":         contents["a.go"],
		"b_renamed.go": contents["b.go"],
		"c.go":         contents["c.go"],
	}
	h4 := ComputeMerkle(filesAlt, contentsAlt)
	if h4 == h1 {
		t.Error("changed file path did not change the Merkle root")
	}
}

// TestMerkleEmpty verifies that an empty file list returns an empty string.
func TestMerkleEmpty(t *testing.T) {
	h := ComputeMerkle(nil, nil)
	if h != "" {
		t.Errorf("expected empty string for empty input, got %q", h)
	}

	h2 := ComputeMerkle([]string{}, map[string][]byte{})
	if h2 != "" {
		t.Errorf("expected empty string for empty slice, got %q", h2)
	}
}

// TestMerkleSingleFile ensures a single-file tree works correctly.
func TestMerkleSingleFile(t *testing.T) {
	files := []string{"main.go"}
	contents := map[string][]byte{"main.go": []byte("package main")}

	h := ComputeMerkle(files, contents)
	if h == "" {
		t.Fatal("expected non-empty hash for single file")
	}
	if h != ComputeMerkle(files, contents) {
		t.Error("single-file hash is not deterministic")
	}
}

// repoRoot returns the top-level directory of the git repository that
// contains the running test binary.
func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("git", "-C", wd, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	root := string(out)
	// Trim trailing newline.
	if len(root) > 0 && root[len(root)-1] == '\n' {
		root = root[:len(root)-1]
	}
	return root, nil
}
