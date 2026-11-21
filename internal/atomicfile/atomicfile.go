// Package atomicfile writes files so that readers never see them half
// written: the data goes to a temporary file in the same directory, which
// then replaces the target in one rename.
package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// Write replaces the file at path with data, creating its directory if
// needed.
func Write(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("atomicfile: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*")
	if err != nil {
		return fmt.Errorf("atomicfile: %w", err)
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("atomicfile: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("atomicfile: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("atomicfile: %w", err)
	}
	return nil
}
