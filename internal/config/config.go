package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/almeidazs/tchess/internal/atomicfile"
)

// DirName is the folder, under the home directory, that holds the
// configuration and the generated sound files.
const DirName = ".tchess"

// Config holds the preferences that persist between runs.
type Config struct {
	// Theme is the board theme's name (see the theme list in the TUI).
	Theme string `toml:"theme"`
	// Sound turns move sounds on or off.
	Sound bool `toml:"sound"`
	// Flip turns the board after every move to face the side to move.
	Flip bool `toml:"flip"`
	// TimeControl is the last time control played ("5+3", "none"); the
	// menu opens on it.
	TimeControl string `toml:"time_control"`
}

// Default returns the preferences used when there is no file yet.
func Default() Config {
	return Config{Sound: true, Flip: true}
}

// Dir returns ~/.tchess, which holds config.toml and the sound files.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: %w", err)
	}
	return filepath.Join(home, DirName), nil
}

// Load reads the file at path. A missing file is not an error: it yields
// the defaults. Keys absent from the file keep their default values.
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("config: %w", err)
	}
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return Default(), fmt.Errorf("config: %s: %w", path, err)
	}
	return cfg, nil
}

// Save writes cfg to path, creating its directory if needed. The file is
// replaced atomically, so a crash never leaves it half written.
func Save(path string, cfg Config) error {
	var buf bytes.Buffer
	buf.WriteString("# tchess preferences. The game keeps this file up to date.\n\n")
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if err := atomicfile.Write(path, buf.Bytes()); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	return nil
}
