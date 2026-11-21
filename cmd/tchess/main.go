// Command tchess is a two-player chess game for the terminal.
//
//	tchess                  open the menu
//	tchess --time 5+3       start a 5 min + 3 s game right away (--time none: no clock)
//	tchess --theme Ocean    use a board theme for this run
//	tchess --no-sound       play without sound for this run
//	tchess --pgn game.pgn   open a game from a PGN file
//	tchess --computer hard  play the computer (needs a UCI engine such as Stockfish)
//	tchess --color black    your side against the computer
//	tchess perft 5 [fen]    count move-tree nodes (move generator self-test)
//	tchess --version        print the version
//
// Preferences (theme, sound, board turning, last time control) live in
// ~/.tchess/config.toml and are saved whenever they change in the game.
// Flags override them for one run without saving. The unfinished game is
// kept in ~/.tchess/current.pgn and exported games go to ~/.tchess/games.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/config"
	"github.com/almeidazs/tchess/internal/sound"
	"github.com/almeidazs/tchess/tui"
	"github.com/almeidazs/tchess/uci"
)

// version is set at build time: -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "perft" {
		if err := runPerft(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return
	}

	timeFlag := flag.String("time", "", `time control "min+sec" (e.g. 5+3) or "none"; without it the menu opens`)
	themeFlag := flag.String("theme", "", "board theme for this run: "+strings.Join(tui.ThemeNames(), ", "))
	noSound := flag.Bool("no-sound", false, "play without sound this run")
	pgnFlag := flag.String("pgn", "", "open a game from a PGN file")
	engineFlag := flag.String("engine", "", "UCI engine to play against (default: stockfish from the PATH)")
	computerFlag := flag.String("computer", "", "play the computer: beginner, easy, medium, hard, expert, max or an Elo")
	colorFlag := flag.String("color", "", "your side against the computer: white, black or random")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("tchess", version)
		return
	}

	opts := tui.Options{Theme: *themeFlag, Mute: *noSound, Config: config.Default()}
	if *timeFlag != "" {
		tc, err := clock.Parse(*timeFlag)
		if err != nil {
			fail(err)
		}
		opts.TimeControl = &tc
	}
	if *themeFlag != "" && !tui.HasTheme(*themeFlag) {
		fail(fmt.Errorf("unknown theme %q (options: %s)", *themeFlag, strings.Join(tui.ThemeNames(), ", ")))
	}
	if *computerFlag != "" {
		if _, ok := tui.ParseLevel(*computerFlag); !ok {
			fail(fmt.Errorf("unknown computer level %q (beginner, easy, medium, hard, expert, max or an Elo)", *computerFlag))
		}
		opts.Computer = *computerFlag
	}
	switch *colorFlag {
	case "", "white", "black", "random":
		opts.Color = *colorFlag
	default:
		fail(fmt.Errorf("unknown color %q (white, black or random)", *colorFlag))
	}
	if *pgnFlag != "" {
		data, err := os.ReadFile(*pgnFlag)
		if err != nil {
			fail(err)
		}
		games, err := chess.ParsePGNGames(string(data))
		if len(games) == 0 {
			fail(fmt.Errorf("%s: no valid game: %w", *pgnFlag, err))
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: games skipped:", err)
		}
		opts.Import, opts.ImportPath = games, *pgnFlag
	}

	// Without a home directory the game still runs, just without saving
	// preferences, keeping games or playing sounds.
	if dir, err := config.Dir(); err == nil {
		path := filepath.Join(dir, "config.toml")
		cfg, err := config.Load(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning:", err, "(using the defaults)")
		}
		opts.Config = cfg
		opts.Save = func(c config.Config) error { return config.Save(path, c) }
		opts.Sound = sound.NewSystem(filepath.Join(dir, "sounds"))
		opts.GamesDir = filepath.Join(dir, "games")
		opts.ResumePath = filepath.Join(dir, "current.pgn")
	}

	engine := opts.Config.Engine
	if *engineFlag != "" {
		engine = *engineFlag
	}
	if rest, ok := strings.CutPrefix(engine, "~/"); ok { // the config is not read by a shell
		if home, err := os.UserHomeDir(); err == nil {
			engine = filepath.Join(home, rest)
		}
	}
	switch path, err := uci.Find(engine); {
	case err == nil:
		opts.EnginePath = path
	case *engineFlag != "":
		fail(fmt.Errorf("engine %q not found", *engineFlag))
	case engine != "":
		fmt.Fprintf(os.Stderr, "warning: engine %q from the config not found\n", engine)
	}
	if opts.Computer != "" && opts.EnginePath == "" {
		fail(errors.New("--computer needs a UCI engine: install stockfish or pass --engine"))
	}

	model := tui.New(opts)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	model.Close() // ends the engine process
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(2)
}
