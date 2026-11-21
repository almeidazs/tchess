// Package sound synthesizes the game's sound effects in code (sine tones
// with envelopes and filtered noise, encoded as WAV) and plays them through
// the operating system's audio player, so the binary needs no cgo and no
// audio files ship with it.
package sound
