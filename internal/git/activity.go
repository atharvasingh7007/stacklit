package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// FileActivity records how many commits touched a file in a time window.
type FileActivity struct {
	Path       string `json:"path"`
	Commits90d int    `json:"commits_90d"`
}

// Activity summarises recent git churn for a repository.
type Activity struct {
	HotFiles    []FileActivity `json:"hot_files"`
	RecentFiles []string       `json:"recent"`
	StableFiles []string       `json:"stable"`
}

// GetActivity analyses git history in root over the last days days.
// If git is unavailable or root is not a git repository, it returns an empty
// Activity without an error.
func GetActivity(root string, days int) (*Activity, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return &Activity{}, nil
	}

	since := fmt.Sprintf("%d.days.ago", days)
	counts, err := countFileCommits(root, since)
	if err != nil {
		// Not a git repo or other git error — return empty gracefully.
		return &Activity{}, nil
	}

	// Build sorted slice of FileActivity (descending by commit count).
	type entry struct {
		path  string
		count int
	}
	entries := make([]entry, 0, len(counts))
	for p, c := range counts {
		entries = append(entries, entry{p, c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].path < entries[j].path
	})

	// HotFiles: top 20.
	hot := make([]FileActivity, 0, 20)
	for i, e := range entries {
		if i >= 20 {
			break
		}
		hot = append(hot, FileActivity{Path: e.path, Commits90d: e.count})
	}

	// StableFiles: 0–1 commits, sample up to 10.
	stable := make([]string, 0, 10)
	for _, e := range entries {
		if e.count <= 1 {
			stable = append(stable, e.path)
			if len(stable) >= 10 {
				break
			}
		}
	}

	// RecentFiles: last 10 unique files from git log -10.
	recent, err := recentFiles(root)
	if err != nil {
		recent = nil
	}

	return &Activity{
		HotFiles:    hot,
		RecentFiles: recent,
		StableFiles: stable,
	}, nil
}

// countFileCommits returns a map of file path → commit count for the window.
func countFileCommits(root, since string) (map[string]int, error) {
	cmd := exec.Command(
		"git", "-C", root,
		"log", "--since="+since,
		"--name-only", "--pretty=format:",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		counts[line]++
	}
	return counts, nil
}

// recentFiles returns up to 10 unique files from the last 10 commits.
func recentFiles(root string) ([]string, error) {
	cmd := exec.Command(
		"git", "-C", root,
		"log", "-10",
		"--name-only", "--pretty=format:",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	result := make([]string, 0, 10)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, ok := seen[line]; !ok {
			seen[line] = struct{}{}
			result = append(result, line)
			if len(result) >= 10 {
				break
			}
		}
	}
	return result, nil
}
