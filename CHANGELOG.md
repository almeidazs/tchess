# Changelog

Every release's changes are listed here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow [Semantic Versioning](https://semver.org/). The notes of each GitHub release are generated from the commits.

## [Unreleased]

First release.

### Added

- Bitboard rules engine: legal moves, castling, en passant, promotion, check, checkmate and stalemate, verified with perft.
- Every FIDE game ending: insufficient material, fivefold repetition and the 75-move rule end the game automatically; threefold repetition and the 50-move rule can be claimed; plus agreement, resignation and flag fall.
- Terminal interface with pixel-art pieces in three sizes, a board that turns to face the player to move (optional) and an animation for every move.
- Mouse (drag or click) and keyboard (arrows that only stop on playable squares), plus moves typed after `:` in SAN or coordinates.
- Fischer-increment clock, time spent on each move and captured pieces.
- Six board themes with a preview in the menu, sounds synthesized by the program itself and a help card with every shortcut on `?`.
- Taking moves back with the opponent's consent, one at a time or back to a position picked while reviewing.
- Review of any earlier position, from the move list or with `[` and `]`.
- PGN: finished games saved automatically, export with `[%clk]` and `[%emt]`, import of files with one or several games, and a choice of time control to continue an unfinished game.
- Resuming an interrupted game, with the clocks as they were.
- Preferences in `~/.tchess/config.toml`.
- deb, rpm, apk and Arch packages and a Docker image at `ghcr.io/almeidazs/tchess`.

[Unreleased]: https://github.com/almeidazs/tchess/commits/main
