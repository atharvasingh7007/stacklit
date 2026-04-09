package git

import (
	"os"
	"path/filepath"
	"strings"
)

const hookScript = `#!/bin/sh
stacklit generate --quiet 2>/dev/null
git add stacklit.json stacklit.mmd 2>/dev/null || true
`

// InstallHook installs (or appends to) the post-commit git hook in root so
// that stacklit regenerates its output on every commit.
//
//   - If .git/hooks/ does not exist, an error is returned.
//   - If the hook already contains "stacklit", it is left unchanged.
//   - If the hook exists but lacks "stacklit", the script is appended.
//   - If the hook does not exist, it is created with mode 0755.
func InstallHook(root string) error {
	hooksDir := filepath.Join(root, ".git", "hooks")
	if _, err := os.Stat(hooksDir); err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, "post-commit")

	existing, err := os.ReadFile(hookPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		// Hook file exists.
		if strings.Contains(string(existing), "stacklit") {
			// Already installed — nothing to do.
			return nil
		}
		// Append to existing hook.
		f, err := os.OpenFile(hookPath, os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString("\n" + hookScript)
		return err
	}

	// Create a new hook file.
	return os.WriteFile(hookPath, []byte(hookScript), 0755)
}
