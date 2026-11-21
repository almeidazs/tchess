package chess

import (
	"errors"
	"strings"
	"testing"
)

func TestPGNRoundTrip(t *testing.T) {
	g := NewGame()
	playSAN(t, g, "e4", "d5", "e5", "f5", "exf6", "Nc6", "fxg7", "Bd7", "gxh8=Q", "e6", "Nf3", "Qe7", "d3", "O-O-O")
	comments := make([]string, g.Ply())
	comments[0] = "[%clk 0:05:00] best by test"
	comments[5] = "a } brace"
	tags := []Tag{{"Event", `Quote "here"`}, {"White", "Alice"}, {"Black", "Bob"}, {"Result", "?"}}
	text := g.PGN(tags, comments)

	if !strings.Contains(text, `[Event "Quote \"here\""]`) || !strings.Contains(text, `[Result "*"]`) {
		t.Fatalf("tags:\n%s", text)
	}
	if !strings.Contains(text, "1. e4 {[%clk 0:05:00] best by test} 1... d5") {
		t.Fatalf("a comment after White's move needs a black move number:\n%s", text)
	}
	for _, line := range strings.Split(text, "\n") {
		if len(line) > pgnLineWidth {
			t.Fatalf("line longer than %d: %q", pgnLineWidth, line)
		}
	}

	pg, err := ParsePGN(text)
	if err != nil {
		t.Fatal(err)
	}
	if pg.Tag("Event") != `Quote "here"` || pg.Tag("White") != "Alice" {
		t.Fatalf("tags %v", pg.Tags)
	}
	got, want := pg.Game.History(), g.History()
	if len(got) != len(want) {
		t.Fatalf("%d moves, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].SAN != want[i].SAN {
			t.Fatalf("move %d: %s, want %s", i, got[i].SAN, want[i].SAN)
		}
	}
	if pg.Comments[0] != "[%clk 0:05:00] best by test" || pg.Comments[5] != "a ) brace" || pg.Comments[1] != "" {
		t.Fatalf("comments %q", pg.Comments)
	}
	if pg.Game.IsOver() || pg.Result != Ongoing {
		t.Fatal("an unfinished game should stay unfinished")
	}
}

func TestPGNCustomStart(t *testing.T) {
	g, err := NewGameFromFEN("4k3/8/8/8/8/8/4P3/4K3 b - - 3 40")
	if err != nil {
		t.Fatal(err)
	}
	playSAN(t, g, "Kd7", "e4")
	text := g.PGN(nil, nil)
	if !strings.Contains(text, `[SetUp "1"]`) || !strings.Contains(text, "40... Kd7 41. e4 *") {
		t.Fatalf("custom start:\n%s", text)
	}
	pg, err := ParsePGN(text)
	if err != nil {
		t.Fatal(err)
	}
	got, want := pg.Game.Position(), g.Position()
	if got.FEN() != want.FEN() {
		t.Fatalf("got %s, want %s", got.FEN(), want.FEN())
	}
}

func TestParsePGNFromOtherPrograms(t *testing.T) {
	text := "\ufeff[Event \"Rated Blitz game\"]\r\n[Site \"https://lichess.org/abc\"]\r\n[Result \"0-1\"]\r\n\r\n" +
		"1.e4 $1 e5 2. Nf3 (2. Nc3 {Vienna} Nf6 (2... Nc6)) 2... Nc6 ; line comment\n" +
		"3. Bc4!? Nf6?! 4. g1g3?? {not legal} 0-1\n"
	if _, err := ParsePGN(text); err == nil || !strings.Contains(err.Error(), "move 7") {
		t.Fatalf("expected an illegal-move error at move 7, got %v", err)
	}

	text = strings.Replace(text, "4. g1g3?? {not legal}", "4. Ng5 d5 5. exd5 Nxd5 6. Nxf7 {fried liver}", 1)
	pg, err := ParsePGN(text)
	if err != nil {
		t.Fatal(err)
	}
	if pg.Game.Ply() != 11 {
		t.Fatalf("%d plies", pg.Game.Ply())
	}
	if pg.Comments[3] != "line comment" || pg.Comments[10] != "fried liver" {
		t.Fatalf("comments %q", pg.Comments)
	}
	if pg.Game.IsOver() || pg.Result != BlackWins {
		t.Fatalf("the moves do not end the game; the result is only recorded: %v / %v", pg.Game.Outcome(), pg.Result)
	}
	if err := pg.Game.Record(pg.Result, Resignation); err != nil || pg.Game.Termination() != Resignation {
		t.Fatalf("record: %v", err)
	}
}

func TestParsePGNMateKeepsRuleTermination(t *testing.T) {
	pg, err := ParsePGN("1. f3 e5 2. g4 Qh4# 0-1")
	if err != nil {
		t.Fatal(err)
	}
	if pg.Game.Termination() != Checkmate {
		t.Fatalf("termination %v", pg.Game.Termination())
	}
	text := pg.Game.PGN(nil, nil)
	if !strings.Contains(text, `[Result "0-1"]`) || !strings.HasSuffix(strings.TrimSpace(text), "Qh4# 0-1") {
		t.Fatalf("export:\n%s", text)
	}
}

func TestParsePGNErrors(t *testing.T) {
	for _, text := range []string{
		`[Event "x]`,
		"1. e4 {open",
		"1. e4 (1. d4",
		"[FEN \"bad\"]\n\n1. e4",
		"1. f3 e5 2. g4 Qh4# 3. a3",
	} {
		if _, err := ParsePGN(text); err == nil {
			t.Errorf("%q: expected an error", text)
		}
	}
}

func TestUndo(t *testing.T) {
	g := NewGame()
	if err := g.Undo(); !errors.Is(err, ErrNothingToUndo) {
		t.Fatalf("undo at start: %v", err)
	}
	playSAN(t, g, "e4", "d5", "exd5", "c6", "dxc6", "Qd6", "cxb7", "Kd8", "bxa8=Q")
	before := g.PositionAt(4)
	for g.Ply() > 4 {
		if err := g.Undo(); err != nil {
			t.Fatal(err)
		}
	}
	if now := g.Position(); now != before || len(g.LegalMoves()) == 0 || len(g.Captured(White)) != 1 {
		t.Fatalf("after undo: %s", now.FEN())
	}
	if g.RepetitionCount() != 1 {
		t.Fatal("repetition keys out of sync after undo")
	}
	playSAN(t, g, "Nf3", "Nf6", "Ng1", "Ng8", "Nf3", "Nf6", "Ng1", "Ng8")
	if g.ClaimableDraw() != ThreefoldRepetition {
		t.Fatal("threefold after undo + replay")
	}
	g.Undo()
	if g.ClaimableDraw() != NoTermination {
		t.Fatal("undo should take back the repetition")
	}

	g = NewGame()
	playSAN(t, g, "f3", "e5", "g4", "Qh4#")
	if err := g.Undo(); !errors.Is(err, ErrGameOver) {
		t.Fatalf("undo after mate: %v", err)
	}
}

func TestRecord(t *testing.T) {
	for _, tt := range []struct {
		o    Outcome
		t    Termination
		good bool
	}{
		{Draw, DrawAgreement, true},
		{WhiteWins, DrawAgreement, false},
		{BlackWins, Resignation, true},
		{Draw, Resignation, false},
		{WhiteWins, Timeout, true},
		{Draw, TimeoutVsInsufficientMaterial, true},
		{Draw, RecordedResult, true},
		{Ongoing, RecordedResult, false},
		{WhiteWins, Checkmate, false}, // only the position can mate
	} {
		g := NewGame()
		err := g.Record(tt.o, tt.t)
		if (err == nil) != tt.good {
			t.Errorf("Record(%v, %v): %v", tt.o, tt.t, err)
		}
		if tt.good && (g.Outcome() != tt.o || g.Termination() != tt.t) {
			t.Errorf("Record(%v, %v) gave %v / %v", tt.o, tt.t, g.Outcome(), g.Termination())
		}
	}
	g := NewGame()
	_ = g.Resign(White)
	if err := g.Record(Draw, DrawAgreement); !errors.Is(err, ErrGameOver) {
		t.Fatalf("recording a finished game: %v", err)
	}
}

func TestPositionAt(t *testing.T) {
	g := NewGame()
	playSAN(t, g, "e4", "e5", "Nf3")
	start := g.StartPosition()
	if p := g.PositionAt(0); p != start || p.FEN() != StartFEN {
		t.Fatal("ply 0 should be the start")
	}
	if p := g.PositionAt(99); p != g.Position() {
		t.Fatal("out-of-range ply should clamp to the current position")
	}
	if p := g.PositionAt(2); p.PieceAt(E5) != MakePiece(Black, Pawn) || p.PieceAt(G1) == NoPiece {
		t.Fatalf("ply 2: %s", p.FEN())
	}
}

func TestParsePGNGames(t *testing.T) {
	text := `[Event "One"]
[Result "1-0"]

1. e4 e5 2. Qh5 Nc6 3. Bc4 Nf6 4. Qxf7# 1-0

[Event "Two"]
[Result "*"]

1. d4 {a comment that wraps
[%clk 0:05:00] onto a line starting with a bracket} d5 ; and { here
2. c4 *

[Event "Broken"]

1. e4 e4 *

[Event "Four"]

1. Nf3 *
`
	games, err := ParsePGNGames(text)
	if len(games) != 3 {
		t.Fatalf("%d games, want 3 (the broken one skipped)", len(games))
	}
	if err == nil || !strings.Contains(err.Error(), "game 3") {
		t.Fatalf("error should name game 3: %v", err)
	}
	for i, want := range []struct {
		event string
		ply   int
	}{{"One", 7}, {"Two", 3}, {"Four", 1}} {
		if games[i].Tag("Event") != want.event || games[i].Game.Ply() != want.ply {
			t.Errorf("game %d: %q with %d plies", i, games[i].Tag("Event"), games[i].Game.Ply())
		}
	}
	if !strings.Contains(games[1].Comments[0], "[%clk 0:05:00]") {
		t.Fatalf("a bracket line inside a comment is not a new game: %q", games[1].Comments)
	}

	first, err := ParsePGN(text)
	if err != nil || first.Tag("Event") != "One" {
		t.Fatalf("ParsePGN reads the first game: %v", err)
	}
	if games, err := ParsePGNGames(""); len(games) != 0 || err != nil {
		t.Fatalf("empty text: %v %v", games, err)
	}
}
