package tui

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/config"
)

func TestPGNClockHelpers(t *testing.T) {
	for _, tt := range []struct {
		tag  string
		want clock.TimeControl
		ok   bool
	}{
		{"300+3", clock.TimeControl{Base: 5 * time.Minute, Increment: 3 * time.Second}, true},
		{"600", clock.TimeControl{Base: 10 * time.Minute}, true},
		{"30+0.5", clock.TimeControl{Base: 30 * time.Second, Increment: 500 * time.Millisecond}, true},
		{"-", clock.TimeControl{}, true},
		{"x", clock.TimeControl{}, false},
		{"", clock.TimeControl{}, false},
	} {
		got, ok := parsePGNTimeControl(tt.tag)
		if got != tt.want || ok != tt.ok {
			t.Errorf("%q: %v %v", tt.tag, got, ok)
		}
	}
	if s := pgnTimeControl(clock.TimeControl{Base: 3 * time.Minute, Increment: 2 * time.Second}); s != "180+2" {
		t.Errorf("pgnTimeControl %q", s)
	}
	if s := formatPGNClock(4*time.Minute + 59*time.Second + 340*time.Millisecond); s != "0:04:59.3" {
		t.Errorf("formatPGNClock %q", s)
	}
	if d, ok := commentDuration("[%clk 0:04:59.3] [%emt 0:00:02]", "%emt"); !ok || d != 2*time.Second {
		t.Errorf("commentDuration %v %v", d, ok)
	}
	if d, ok := commentDuration("[%clk 1:02:03]", "%clk"); !ok || d != time.Hour+2*time.Minute+3*time.Second {
		t.Errorf("commentDuration %v %v", d, ok)
	}
	if _, ok := commentDuration("nothing", "%clk"); ok {
		t.Error("no clock in a plain comment")
	}
}

func TestExportWritesPGN(t *testing.T) {
	dir := t.TempDir()
	clk := &fakeTime{time.Date(2026, 11, 21, 20, 30, 5, 0, time.UTC)}
	m := newDiskModel(t, dir, clk, config.Default())
	m.startGame(clock.TimeControl{Base: 3 * time.Minute, Increment: 2 * time.Second})
	moveRun(m, chess.E2, chess.E4)
	clk.advance(4 * time.Second)
	moveRun(m, chess.E7, chess.E5)
	press(m, "e")

	path := filepath.Join(dir, "games", "2026-11-21_20-30-09.pgn")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pg, err := chess.ParsePGN(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if pg.Game.Ply() != 2 || pg.Tag("TimeControl") != "180+2" || pg.Tag("Date") != "2026.11.21" {
		t.Fatalf("exported:\n%s", data)
	}
	if !strings.Contains(pg.Comments[1], "[%clk 0:02:58.0]") || !strings.Contains(pg.Comments[1], "[%emt 0:00:04.0]") {
		t.Fatalf("clock comments %q", pg.Comments)
	}
	if !strings.Contains(m.notice, "Game saved in") {
		t.Fatalf("notice %q", m.notice)
	}
}

func TestResumeAfterRestart(t *testing.T) {
	dir := t.TempDir()
	clk := &fakeTime{time.Unix(1000, 0)}
	m := newDiskModel(t, dir, clk, config.Default())
	runCmd(m, m.startGame(clock.TimeControl{Base: 5 * time.Minute, Increment: 3 * time.Second}))
	moveRun(m, chess.E2, chess.E4)
	clk.advance(5 * time.Second)
	moveRun(m, chess.E7, chess.E5)
	clk.advance(7 * time.Second)
	moveRun(m, chess.G1, chess.F3)
	wantPos := m.game.Position()

	// A new run finds the game and offers it first.
	m2 := newDiskModel(t, dir, clk, config.Default())
	if m2.resume == nil || m2.menuEntries()[0].kind != entryResume || m2.menu.cursor != 0 {
		t.Fatal("the menu should offer the saved game first")
	}
	if !strings.Contains(m2.menuView(), "5+3 · move 2") {
		t.Fatalf("menu:\n%s", m2.menuView())
	}
	press(m2, "enter")
	if m2.screen != screenGame || m2.game.Position() != wantPos || m2.tc.Increment != 3*time.Second {
		t.Fatalf("resumed: screen %v ply %d", m2.screen, m2.game.Ply())
	}
	if !m2.clock.Running() {
		t.Fatal("the side to move's clock runs at once")
	}
	if got := m2.clock.Remaining(chess.White); got != 5*time.Minute-7*time.Second+3*time.Second {
		t.Fatalf("white %v", got)
	}
	if got := m2.clock.Remaining(chess.Black); got != 5*time.Minute-5*time.Second+3*time.Second {
		t.Fatalf("black %v", got)
	}
	if len(m2.moveTimes) != 3 || m2.moveTimes[2] != 7*time.Second {
		t.Fatalf("move times %v", m2.moveTimes)
	}
	if m2.orientation != chess.Black {
		t.Fatal("the board faces the side to move")
	}
}

func TestFinishedOrNewGameForgetsSave(t *testing.T) {
	dir := t.TempDir()
	clk := &fakeTime{time.Unix(0, 0)}
	path := filepath.Join(dir, "current.pgn")
	m := newDiskModel(t, dir, clk, config.Default())
	runCmd(m, m.startGame(clock.TimeControl{}))
	moveRun(m, chess.F2, chess.F3)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the game should be kept after a move: %v", err)
	}
	moveRun(m, chess.E7, chess.E5)
	moveRun(m, chess.G2, chess.G4)
	moveRun(m, chess.D8, chess.H4) // mate
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a finished game has nothing to resume: %v", err)
	}

	runCmd(m, m.startGame(clock.TimeControl{}))
	moveRun(m, chess.E2, chess.E4)
	m2 := newDiskModel(t, dir, clk, config.Default())
	if m2.resume == nil {
		t.Fatal("saved game not found")
	}
	press(m2, "down")
	press(m2, "down") // past "Continue" and "Opponent", the first time control
	press(m2, "enter")
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) || m2.resume != nil {
		t.Fatalf("starting another game discards the saved one: %v", err)
	}
}

func TestFinishedGamesAreSaved(t *testing.T) {
	dir := t.TempDir()
	clk := &fakeTime{time.Date(2026, 11, 21, 21, 0, 0, 0, time.UTC)}
	m := newDiskModel(t, dir, clk, config.Default())
	runCmd(m, m.startGame(clock.TimeControl{}))
	moveRun(m, chess.E2, chess.E4)
	press(m, "d")
	press(m, "n") // a refused draw leaves a notice that must not linger
	moveRun(m, chess.E7, chess.E5)
	press(m, "r")
	press(m, "y") // White, to move, resigns

	files, err := filepath.Glob(filepath.Join(dir, "games", "*.pgn"))
	if err != nil || len(files) != 1 {
		t.Fatalf("finished games go to the games folder: %v %v", files, err)
	}
	pg, err := readPGNFile(files[0])
	if err != nil || pg.Game.Ply() != 2 || pg.Result != chess.BlackWins || pg.Tag("Termination") != "resignation" {
		t.Fatalf("saved game: %v %+v", err, pg.Tags)
	}
	if status := m.statusView(); !strings.Contains(status, "Game saved in") || strings.Contains(status, "declined") {
		t.Fatalf("status %q", status)
	}
	press(m, "e")
	if again, _ := filepath.Glob(filepath.Join(dir, "games", "*.pgn")); len(again) != 1 || !strings.Contains(m.notice, "already saved") {
		t.Fatalf("e after the end should not save a copy: %v %q", again, m.notice)
	}

	// The picker lists it.
	press(m, "n")
	m.openPicker()
	if len(m.open.files) != 1 || !strings.Contains(m.View(), "2026-11-21_21-00-00") {
		t.Fatalf("picker files %v", m.open.files)
	}
}

func TestReopenedGameKeepsItsEnding(t *testing.T) {
	dir := t.TempDir()
	clk := &fakeTime{time.Unix(0, 0)}
	m := newDiskModel(t, dir, clk, config.Default())
	runCmd(m, m.startGame(clock.TimeControl{}))
	moveRun(m, chess.G2, chess.G3)
	moveRun(m, chess.C7, chess.C6)
	press(m, "d")
	press(m, "y") // draw agreed

	press(m, "n")
	m.openPicker()
	press(m, "enter")
	if m.mode != modeGameOver || m.game.Termination() != chess.DrawAgreement {
		t.Fatalf("reopened: mode %v termination %v", m.mode, m.game.Termination())
	}
	if !strings.Contains(m.statusView(), "agreement") {
		t.Fatalf("status %q", m.statusView())
	}
	if !m.resultHidden {
		t.Fatal("an opened game shows the board, not the result card")
	}
	press(m, "esc")
	if m.resultHidden {
		t.Fatal("esc brings the card back")
	}
	press(m, "esc")
	if !m.resultHidden || m.screen != screenGame {
		t.Fatal("esc puts it aside again, without leaving")
	}
	press(m, "e")
	if files, _ := filepath.Glob(filepath.Join(dir, "games", "*.pgn")); len(files) != 1 || !strings.Contains(m.notice, "already saved") {
		t.Fatalf("e on a reopened game should not copy it: %v %q", files, m.notice)
	}
}

func TestArrowsStepThroughFinishedGame(t *testing.T) {
	m := newGame(t, 140, 50, "none")
	for _, sq := range []chess.Square{chess.F2, chess.F3, chess.E7, chess.E5, chess.G2, chess.G4, chess.D8, chess.H4} {
		click(m, sq)
	}
	if m.resultHidden {
		t.Fatal("a game that just ended shows its result")
	}
	key(m, "left")
	key(m, "left")
	if !m.reviewing || m.reviewPly != 2 {
		t.Fatalf("← steps back: reviewing %v ply %d", m.reviewing, m.reviewPly)
	}
	key(m, "right")
	if m.reviewPly != 3 {
		t.Fatalf("→ steps forward: ply %d", m.reviewPly)
	}
	key(m, "esc")
	if m.reviewing || m.resultHidden || m.screen != screenGame {
		t.Fatal("esc while reviewing goes back to the end, not out of the game")
	}
}

func TestImportedFinishedGameIsNotSavedAgain(t *testing.T) {
	dir := t.TempDir()
	pg, err := chess.ParsePGN("1. f3 e5 2. g4 Qh4# 0-1")
	if err != nil {
		t.Fatal(err)
	}
	m := New(Options{Config: config.Default(), Import: []*chess.PGNGame{pg}, GamesDir: filepath.Join(dir, "games")})
	m.Update(teaSize(140, 50))
	runCmd(m, m.Init())
	if files, _ := filepath.Glob(filepath.Join(dir, "games", "*.pgn")); len(files) != 0 {
		t.Fatalf("opening a finished game should not copy it: %v", files)
	}
}

func TestImportFinishedGame(t *testing.T) {
	pg, err := chess.ParsePGN("[Result \"1-0\"]\n\n1. e4 e5 2. Qh5 Nc6 3. Bc4 Nf6 1-0")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	m := New(Options{Config: config.Default(), Import: []*chess.PGNGame{pg}, ResumePath: filepath.Join(dir, "current.pgn")})
	m.Update(teaSize(140, 50))
	runCmd(m, m.Init())
	if m.screen != screenGame || m.mode != modeGameOver || m.game.Termination() != chess.RecordedResult {
		t.Fatalf("screen %v mode %v", m.screen, m.mode)
	}
	if !strings.Contains(m.View(), "result from the PGN") {
		t.Fatal("the result card explains where the result came from")
	}
	key(m, "home")
	if !m.reviewing {
		t.Fatal("an imported game can be reviewed")
	}
	if _, err := os.Stat(filepath.Join(dir, "current.pgn")); err == nil {
		t.Fatal("a finished import is not a game to resume")
	}
}

func TestImportUnfinishedAsksForClocks(t *testing.T) {
	text := "[TimeControl \"300+3\"]\n\n1. e4 {[%clk 0:05:00]} e5 {[%clk 0:04:50]} *"
	open := func(t *testing.T) (*Model, *fakeTime) {
		t.Helper()
		pg, err := chess.ParsePGN(text)
		if err != nil {
			t.Fatal(err)
		}
		clk := &fakeTime{time.Unix(0, 0)}
		m := New(Options{Config: config.Default(), Import: []*chess.PGNGame{pg}, Now: clk.now})
		m.Update(teaSize(140, 50))
		m.animate = false
		runCmd(m, m.Init())
		return m, clk
	}

	// The record's own clocks: White to move with 5:00 left, running.
	m, clk := open(t)
	if m.screen != screenClocks || m.clocks.options[0].src != clocksFromRecord {
		t.Fatalf("screen %v options %v", m.screen, m.clocks.options)
	}
	if !strings.Contains(m.View(), "PGN clocks") || !strings.Contains(m.View(), "5:00 · 4:50") {
		t.Fatalf("view:\n%s", m.View())
	}
	press(m, "enter")
	if m.screen != screenGame || m.tc.Increment != 3*time.Second || !m.clock.Running() {
		t.Fatalf("screen %v tc %v", m.screen, m.tc)
	}
	clk.advance(2 * time.Second)
	if got := m.clock.Remaining(chess.White); got != 4*time.Minute+58*time.Second {
		t.Fatalf("white %v", got)
	}
	if got := m.clock.Remaining(chess.Black); got != 4*time.Minute+50*time.Second {
		t.Fatalf("black %v", got)
	}

	// A fresh time control: both clocks full.
	m, _ = open(t)
	m.clocks.cursor = 2 // Blitz 3+2 (after the record and Bullet)
	press(m, "enter")
	if m.tc.Base != 3*time.Minute || m.clock.Remaining(chess.Black) != 3*time.Minute || m.clockFrom != 2 {
		t.Fatalf("fresh clocks: tc %v black %v from %d", m.tc, m.clock.Remaining(chess.Black), m.clockFrom)
	}
	click(m, chess.G1)
	click(m, chess.F3)
	pgn := m.gamePGN()
	if strings.Count(pgn, "%clk") != 1 {
		t.Fatalf("only the move made with the new clock has a %%clk:\n%s", pgn)
	}

	// No clock, and Esc goes back to the menu.
	m, _ = open(t)
	press(m, "esc")
	if m.screen != screenMenu {
		t.Fatal("esc returns to the menu")
	}
	m, _ = open(t)
	for m.clocks.options[m.clocks.cursor].src != clocksOff {
		press(m, "down")
	}
	press(m, "enter")
	if m.clock != nil || m.mode != modePlay {
		t.Fatal("No clock plays on without a clock")
	}
}

func TestTerminationFromText(t *testing.T) {
	for text, want := range map[string]chess.Termination{
		"agreement":           chess.DrawAgreement,
		"time forfeit":        chess.Timeout,
		"acordo mútuo":        chess.DrawAgreement, // games saved by older versions
		"desistência":         chess.Resignation,
		"repetição tripla":    chess.ThreefoldRepetition,
		"something else":      chess.RecordedResult,
		"":                    chess.RecordedResult,
		"result from the PGN": chess.RecordedResult,
	} {
		if got := terminationFromText(text); got != want {
			t.Errorf("%q: %v, want %v", text, got, want)
		}
	}
}
