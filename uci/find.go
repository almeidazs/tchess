package uci

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// ErrNotFound is returned by Find when no engine is available.
var ErrNotFound = errors.New("uci: no engine found")

// fallbackDirs are searched for "stockfish" when it is not on the PATH:
// Debian and Ubuntu install it in /usr/games, which login shells add to
// the PATH but other launchers may not.
var fallbackDirs = []string{"/usr/games", "/usr/local/games", "/opt/homebrew/bin", "/usr/local/bin"}

// Find returns the engine to run: configured (a path or a program name)
// when given, otherwise Stockfish from the PATH or a usual location.
func Find(configured string) (string, error) {
	if configured != "" {
		path, err := exec.LookPath(configured)
		if err != nil {
			return "", err
		}
		return path, nil
	}
	if path, err := exec.LookPath("stockfish"); err == nil {
		return path, nil
	}
	for _, dir := range fallbackDirs {
		path := filepath.Join(dir, "stockfish")
		if info, err := os.Stat(path); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return path, nil
		}
	}
	return "", ErrNotFound
}
