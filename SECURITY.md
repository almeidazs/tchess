# Security

## Supported versions

Only the latest release gets security fixes. Before reporting, check that the problem happens there (`tchess --version`).

## Reporting a vulnerability

Please don't open a public issue. Use GitHub's private reporting, under [Security → Report a vulnerability](https://github.com/almeidazs/tchess/security/advisories/new), and include:

- the tchess version and your operating system;
- steps to reproduce, with the PGN file or config involved, if any;
- what an attacker could do with it.

You'll get a first answer within 7 days. Once the problem is confirmed, the fix ships in a new release and the advisory is published with it, crediting whoever found it if they want.

## What counts as a vulnerability

tchess doesn't use the network. Outside the terminal, it only:

- reads PGN files the player chooses (from the menu or with `--pgn`);
- reads and writes its config, games and sounds in `~/.tchess`;
- runs the system audio player (`pw-play`, `paplay`, `aplay`, `afplay` or PowerShell) with the path of a WAV file it generated itself.

The interesting cases are a PGN file that makes the program hang, use unbounded memory or crash; a write outside `~/.tchess`; and any way to run commands through a file path. Chess rule mistakes and interface problems are ordinary bugs and can go straight to the [issues](https://github.com/almeidazs/tchess/issues).
