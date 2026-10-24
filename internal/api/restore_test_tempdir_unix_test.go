//go:build unix

package api

import (
	"os"
	"path/filepath"
	"testing"
)

func privilegeDropTestDirs(t *testing.T) (string, string) {
	t.Helper()

	root := t.TempDir()
	replacementDir := filepath.Join(root, "replacement")
	liveDir := filepath.Join(root, "live")
	for _, path := range []string{replacementDir, liveDir} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatalf("create restore test directory: %v", err)
		}
	}

	if os.Geteuid() != 0 {
		return replacementDir, liveDir
	}

	// InitDB intentionally drops root to the runtime UID/GID. Keep only this
	// test's temporary tree traversable and writable after that transition.
	const runtimeID = 65532
	for _, path := range []string{filepath.Dir(root), root, replacementDir, liveDir} {
		if err := os.Chown(path, runtimeID, runtimeID); err != nil {
			t.Fatalf("prepare restore test directory ownership: %v", err)
		}
	}

	return replacementDir, liveDir
}
