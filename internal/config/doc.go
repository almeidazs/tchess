// Package config loads and saves the player's preferences (board theme,
// sound, last time control) in ~/.tchess/config.toml. A missing or partial
// file falls back to the defaults, and saves replace the file atomically.
package config
