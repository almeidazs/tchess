package uci

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/uci/ucitest"
)

func TestMain(m *testing.M) {
	ucitest.Main()
	os.Exit(m.Run())
}

// startFake runs the fake engine (this test binary), logging its input.
func startFake(t *testing.T) (*Engine, string) {
	t.Helper()
	log := filepath.Join(t.TempDir(), "log")
	t.Setenv(ucitest.LogVar, log)
	e, err := Start(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { e.Close() })
	return e, log
}

func readLog(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestHandshake(t *testing.T) {
	e, _ := startFake(t)
	if e.Name != "FakeFish 1" {
		t.Fatalf("name %q", e.Name)
	}
	elo, ok := e.Option("uci_elo")
	if !ok || elo.Type != "spin" || elo.Min != 1320 || elo.Max != 3190 {
		t.Fatalf("UCI_Elo %+v", elo)
	}
	if skill, ok := e.Option("Skill Level"); !ok || skill.Max != 20 {
		t.Fatalf("option with a space in its name: %+v", skill)
	}
}

func TestSearch(t *testing.T) {
	e, log := startFake(t)
	if err := e.SetOption("UCI_Elo", "1600"); err != nil {
		t.Fatal(err)
	}
	if err := e.NewGame(); err != nil {
		t.Fatal(err)
	}
	res, err := e.Search(context.Background(), chess.StartFEN, []string{"e2e4"}, Limits{WTime: time.Minute, BTime: 50 * time.Second, WInc: 2 * time.Second, BInc: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	pos, _ := chess.ParseFEN(chess.StartFEN)
	mv, _ := pos.ParseUCI("e2e4")
	pos.Make(mv)
	if _, err := pos.ParseUCI(res.BestMove); err != nil {
		t.Fatalf("best move %q is not legal for Black: %v", res.BestMove, err)
	}
	if !res.Score.Set || res.Score.CP != 25 || res.Depth != 1 || len(res.PV) != 1 {
		t.Fatalf("result %+v", res)
	}
	got := readLog(t, log)
	for _, want := range []string{
		"setoption name UCI_Elo value 1600",
		"ucinewgame",
		"position fen " + chess.StartFEN + " moves e2e4",
		"go wtime 60000 btime 50000 winc 2000 binc 2000",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("engine did not receive %q:\n%s", want, got)
		}
	}
}

func TestSearchCancelled(t *testing.T) {
	e, log := startFake(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res, err := e.Search(ctx, chess.StartFEN, nil, Limits{})
	if !errors.Is(err, context.Canceled) || res.BestMove == "" {
		t.Fatalf("canceled search: %+v %v", res, err)
	}
	if got := readLog(t, log); !strings.Contains(got, "go infinite") {
		t.Fatalf("no limits means an infinite search:\n%s", got)
	}
	// The engine is still usable afterwards.
	if _, err := e.Search(context.Background(), chess.StartFEN, nil, Limits{Depth: 3}); err != nil {
		t.Fatal(err)
	}
}

func TestStartFailures(t *testing.T) {
	if _, err := Start(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected an error for a missing program")
	}
	// A program that is not an engine exits without "uciok".
	if runtime.GOOS != "windows" {
		if _, err := Start("true"); err == nil {
			t.Fatal("expected an error for a non-engine")
		}
	}
}

func TestParseInfo(t *testing.T) {
	var res Result
	parseInfo("info depth 12 seldepth 18 multipv 1 score mate -3 nodes 1000 pv e7e5 g1f3", &res)
	if res.Score.Mate != -3 || res.Depth != 12 || len(res.PV) != 2 {
		t.Fatalf("%+v", res)
	}
	parseInfo("info depth 13 currmove e2e4 currmovenumber 1", &res)
	if res.Depth != 12 {
		t.Fatal("lines without a score are ignored")
	}
	parseInfo("info depth 14 score cp 31 lowerbound pv d2d4", &res)
	if res.Score.CP != 31 || res.Score.Mate != 0 || res.PV[0] != "d2d4" {
		t.Fatalf("%+v", res)
	}
}

func TestParseOption(t *testing.T) {
	o, ok := parseOption("option name Clear Hash type button")
	if !ok || o.Name != "Clear Hash" || o.Type != "button" {
		t.Fatalf("%+v", o)
	}
	o, _ = parseOption("option name Style type combo default Normal var Solid var Normal var Risky")
	if o.Default != "Normal" {
		t.Fatalf("%+v", o)
	}
}

func TestFind(t *testing.T) {
	if _, err := Find(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("a configured path that does not exist is an error")
	}
	if runtime.GOOS != "windows" { // "true" is a Unix program
		if path, err := Find("true"); err != nil || !strings.HasSuffix(path, "true") {
			t.Fatalf("a configured program name is looked up: %q %v", path, err)
		}
	}
	t.Setenv("PATH", t.TempDir())
	saved := fallbackDirs
	fallbackDirs = nil
	defer func() { fallbackDirs = saved }()
	if _, err := Find(""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("no stockfish anywhere: %v", err)
	}
}
