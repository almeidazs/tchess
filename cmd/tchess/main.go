// Command tchess is a two-player chess game for the terminal.
//
//	tchess                  open the menu
//	tchess --time 5+3       start a 5 min + 3 s game right away (--time none: no clock)
//	tchess --theme Oceano   use a board theme for this run
//	tchess --no-sound       play without sound for this run
//	tchess --pgn game.pgn   open a game from a PGN file
//	tchess perft 5 [fen]    count move-tree nodes (move generator self-test)
//	tchess --version        print the version
//
// Preferences (theme, sound, board turning, last time control) live in
// ~/.tchess/config.toml and are saved whenever they change in the game.
// Flags override them for one run without saving. The unfinished game is
// kept in ~/.tchess/current.pgn and exported games go to ~/.tchess/games.
package main

import (
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
)

// version is set at build time: -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "perft" {
		if err := runPerft(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "erro:", err)
			os.Exit(2)
		}
		return
	}

	timeFlag := flag.String("time", "", `controle de tempo "min+seg" (ex.: 5+3) ou "none"; omitido abre o menu`)
	themeFlag := flag.String("theme", "", "tema do tabuleiro nesta execução: "+strings.Join(tui.ThemeNames(), ", "))
	noSound := flag.Bool("no-sound", false, "joga sem som nesta execução")
	pgnFlag := flag.String("pgn", "", "abre uma partida de um arquivo PGN")
	showVersion := flag.Bool("version", false, "mostra a versão e sai")
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
		fail(fmt.Errorf("tema desconhecido %q (opções: %s)", *themeFlag, strings.Join(tui.ThemeNames(), ", ")))
	}
	if *pgnFlag != "" {
		data, err := os.ReadFile(*pgnFlag)
		if err != nil {
			fail(err)
		}
		pg, err := chess.ParsePGN(string(data))
		if err != nil {
			fail(fmt.Errorf("%s: %w", *pgnFlag, err))
		}
		opts.Import = pg
	}

	// Without a home directory the game still runs, just without saving
	// preferences, keeping games or playing sounds.
	if dir, err := config.Dir(); err == nil {
		path := filepath.Join(dir, "config.toml")
		cfg, err := config.Load(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "aviso:", err, "(usando o padrão)")
		}
		opts.Config = cfg
		opts.Save = func(c config.Config) error { return config.Save(path, c) }
		opts.Sound = sound.NewSystem(filepath.Join(dir, "sounds"))
		opts.GamesDir = filepath.Join(dir, "games")
		opts.ResumePath = filepath.Join(dir, "current.pgn")
	}

	p := tea.NewProgram(tui.New(opts), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "erro:", err)
	os.Exit(2)
}
