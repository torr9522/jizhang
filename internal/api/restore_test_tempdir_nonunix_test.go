//go:build !unix

package api

import "testing"

func privilegeDropTestDirs(t *testing.T) (string, string) {
	t.Helper()
	return t.TempDir(), t.TempDir()
}
