package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/config"
	"github.com/almeidazs/tchess/uci"
	"github.com/almeidazs/tchess/uci/ucitest"
)

// settle runs cmd and every command the resulting messages lead to, so the
// engine's answers arrive before the test goes on. Clock ticks and
// animation frames are dropped, which keeps it from running forever.
func settle(t *testing.T, m *Model, cmd tea.Cmd) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	queue := []tea.Cmd{cmd}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c == nil {
			continue
		}
		done := make(chan tea.Msg, 1)
		go func() { done <- c() }()
		select {
		case msg := <-done:
			switch msg := msg.(type) {
			case tea.BatchMsg:
				queue = append(queue, msg...)
			case tickMsg, animMsg, nil:
			default:
				_, next := m.Update(msg)
				queue = append(queue, next)
			}
		case <-time.After(time.Until(deadline)):
			t.Fatal("the engine did not answer in time")
		}
	}
}

// do sends a message and settles what follows.
func do(t *testing.T, m *Model, msg tea.Msg) {
	t.Helper()
	_, cmd := m.Update(msg)
	settle(t, m, cmd)
}

func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

// clickSq presses and releases the mouse on a square, settling after each.
func clickSq(t *testing.T, m *Model, sq chess.Square) {
	t.Helper()
	x, y := m.SquareCenter(sq)
	do(t, m, tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	do(t, m, tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})
}

// newComputerModel creates a model that plays the fake engine, logging
// what the engine receives.
func newComputerModel(t *testing.T, edit func(*config.Config), dir string) (*Model, string) {
	t.Helper()
	log := filepath.Join(t.TempDir(), "uci.log")
	t.Setenv(ucitest.LogVar, log)
	cfg := config.Default()
	cfg.Opponent, cfg.ComputerElo, cfg.HumanColor = "computer", 1600, "white"
	if edit != nil {
		edit(&cfg)
	}
	opts := Options{Config: cfg, EnginePath: os.Args[0]}
	if dir != "" {
		opts.GamesDir, opts.ResumePath = filepath.Join(dir, "games"), filepath.Join(dir, "current.pgn")
	}
	m := New(opts)
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 50})
	m.animate = false
	t.Cleanup(m.Close)
	return m, log
}

func engineLog(t *testing.T, path string) string {
	t.Helper()
	data, _ := os.ReadFile(path)
	return string(data)
}

func TestComputerAnswersTheHumansMove(t *testing.T) {
	m, log := newComputerModel(t, nil, "")
	settle(t, m, m.startGame(clock.TimeControl{Base: 5 * time.Minute, Increment: 3 * time.Second}))
	if !m.vs.computer || m.vs.human != chess.White || m.thinking {
		t.Fatalf("vs %+v thinking %v", m.vs, m.thinking)
	}
	clickSq(t, m, chess.E2)
	clickSq(t, m, chess.E4)
	if m.game.Ply() != 2 || m.game.Turn() != chess.White || m.thinking {
		t.Fatalf("the computer should have answered: ply %d", m.game.Ply())
	}
	if m.orientation != chess.White {
		t.Fatal("the board stays on the human's side")
	}
	got := engineLog(t, log)
	for _, want := range []string{"setoption name UCI_LimitStrength value true", "setoption name UCI_Elo value 1600", "ucinewgame", "moves e2e4", "go wtime"} {
		if !strings.Contains(got, want) {
			t.Fatalf("engine did not receive %q:\n%s", want, got)
		}
	}
	if !strings.Contains(m.statusView(), "Your move") || !strings.Contains(m.View(), "You") {
		t.Fatalf("status %q", m.statusView())
	}
	if m.engineName != "FakeFish 1" || !strings.Contains(m.View(), "FakeFish") {
		t.Fatal("the panel names the engine")
	}
}

func TestComputerOpensWithWhite(t *testing.T) {
	m, _ := newComputerModel(t, func(c *config.Config) { c.HumanColor = "black" }, "")
	settle(t, m, m.startGame(clock.TimeControl{}))
	if m.game.Ply() != 1 || m.vs.human != chess.Black || m.orientation != chess.Black {
		t.Fatalf("ply %d human %v orientation %v", m.game.Ply(), m.vs.human, m.orientation)
	}
}

func TestHumanWaitsForTheComputer(t *testing.T) {
	m, _ := newComputerModel(t, func(c *config.Config) { c.HumanColor = "black" }, "")
	m.startGame(clock.TimeControl{}) // not settled: the computer is still thinking
	if !m.thinking || !strings.Contains(m.statusView(), "is thinking") {
		t.Fatalf("thinking %v status %q", m.thinking, m.statusView())
	}
	x, y := m.SquareCenter(chess.E7)
	m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if m.selected != chess.NoSquare || !strings.Contains(m.notice, "Wait") {
		t.Fatal("the human cannot pick a piece on the computer's turn")
	}
	m.Update(keyMsg(":"))
	if m.mode == modeInput {
		t.Fatal("nor type a move")
	}
}

func TestUndoAgainstComputerNeedsNoConsent(t *testing.T) {
	dir := t.TempDir()
	m, _ := newComputerModel(t, nil, dir)
	settle(t, m, m.startGame(clock.TimeControl{}))
	clickSq(t, m, chess.E2)
	clickSq(t, m, chess.E4)
	do(t, m, keyMsg("u"))
	if m.game.Ply() != 0 || m.mode != modePlay || !m.assistTakeback {
		t.Fatalf("u should take back both moves at once: ply %d mode %v", m.game.Ply(), m.mode)
	}
	clickSq(t, m, chess.D2)
	clickSq(t, m, chess.D4)
	if pgn := m.gamePGN(); !strings.Contains(pgn, `[TchessAssist "takebacks"]`) || !strings.Contains(pgn, `[White "Player"]`) {
		t.Fatalf("PGN:\n%s", pgn)
	}
}

func TestHint(t *testing.T) {
	m, _ := newComputerModel(t, nil, "")
	settle(t, m, m.startGame(clock.TimeControl{}))
	do(t, m, keyMsg("t"))
	if !m.hintSet || !strings.HasPrefix(m.notice, "Hint: ") || !m.assistHint {
		t.Fatalf("hint %v notice %q", m.hintSet, m.notice)
	}
	bs := m.boardState(computeLayout(m.width, m.height))
	if !bs.marked.Has(m.hint.From()) || !bs.targets.Has(m.hint.To()) {
		t.Fatal("the hint is highlighted")
	}
	if !strings.Contains(m.gamePGN(), "hints") {
		t.Fatal("hints are recorded")
	}
}

func TestDrawOfferToComputer(t *testing.T) {
	for _, tt := range []struct {
		score  string
		accept bool
	}{{"25", true}, {"0", true}, {"-300", false}} {
		t.Setenv(ucitest.ScoreVar, tt.score) // from the human's (side to move) view
		m, _ := newComputerModel(t, nil, "")
		settle(t, m, m.startGame(clock.TimeControl{}))
		do(t, m, keyMsg("d"))
		if got := m.mode == modeGameOver && m.game.Termination() == chess.DrawAgreement; got != tt.accept {
			t.Errorf("score %s: accepted %v, want %v (notice %q)", tt.score, got, tt.accept, m.notice)
		}
	}
}

func TestEvaluationInReview(t *testing.T) {
	m, _ := newComputerModel(t, func(c *config.Config) { c.Opponent = "human" }, "")
	settle(t, m, m.startGame(clock.TimeControl{}))
	clickSq(t, m, chess.E2)
	clickSq(t, m, chess.E4)
	if len(m.evals) != 0 || strings.Contains(m.statusView(), "eval") {
		t.Fatal("no evaluation during the game")
	}
	do(t, m, keyMsg("["))
	e, ok := m.evals[0]
	if !ok || e.cp != 25 { // White to move: the fake's +25 stays +25
		t.Fatalf("evals %v", m.evals)
	}
	if !strings.Contains(m.statusView(), "eval +0.2") {
		t.Fatalf("status %q", m.statusView())
	}
	if bar := m.evalBar(8 * computeLayout(m.width, m.height).cellH); len(bar) == 0 {
		t.Fatal("the bar is drawn")
	}
	do(t, m, keyMsg("]"))
	if m.reviewing || strings.Contains(m.statusView(), "eval") {
		t.Fatal("back to the game: the evaluation hides again")
	}
}

func TestResumeAgainstComputer(t *testing.T) {
	dir := t.TempDir()
	m, _ := newComputerModel(t, func(c *config.Config) { c.ComputerElo = 2000 }, dir)
	settle(t, m, m.startGame(clock.TimeControl{}))
	clickSq(t, m, chess.E2)
	clickSq(t, m, chess.E4)
	m.Close()

	m2, _ := newComputerModel(t, func(c *config.Config) { c.Opponent = "human" }, dir)
	if m2.resume == nil {
		t.Fatal("saved game not found")
	}
	do(t, m2, keyMsg("enter"))
	if !m2.vs.computer || m2.vs.elo != 2000 || m2.vs.human != chess.White || m2.game.Ply() != 2 {
		t.Fatalf("resumed vs %+v ply %d", m2.vs, m2.game.Ply())
	}
}

func TestNoEngine(t *testing.T) {
	cfg := config.Default()
	cfg.Opponent = "computer"
	m := New(Options{Config: cfg})
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 50})
	if !strings.Contains(m.menuView(), "Computer (no engine)") {
		t.Fatalf("menu:\n%s", m.menuView())
	}
	m.openOpponent()
	if !strings.Contains(m.View(), "No UCI engine found") || !strings.Contains(m.View(), "apt install stockfish") {
		t.Fatalf("view:\n%s", m.View())
	}
	m.Update(keyMsg("enter")) // "Computer" is highlighted but disabled
	if m.opp.step != 0 || m.opp.err == "" {
		t.Fatal("the computer cannot be chosen without an engine")
	}
	m.screen = screenMenu
	m.startGame(clock.TimeControl{})
	if m.vs.computer || !strings.Contains(m.notice, "No engine found") {
		t.Fatalf("a saved computer choice falls back to two players: %+v %q", m.vs, m.notice)
	}
}

func TestOpponentPicker(t *testing.T) {
	m, _ := newComputerModel(t, func(c *config.Config) { c.Opponent = "human" }, "")
	rec := &saveRecorder{}
	m.saver = &saver{fn: rec.save}
	m.openOpponent()
	for _, k := range []string{"down", "enter"} { // Computer
		do(t, m, keyMsg(k))
	}
	if m.opp.step != 1 || m.opp.cursor != 2 { // Medium preselected
		t.Fatalf("step %d cursor %d", m.opp.step, m.opp.cursor)
	}
	for range 4 { // to "Custom Elo…"
		do(t, m, keyMsg("down"))
	}
	do(t, m, keyMsg("enter"))
	for _, k := range []string{"1", "7", "5", "0", "enter"} {
		do(t, m, keyMsg(k))
	}
	if m.opp.step != 2 || m.opp.choice.elo != 1750 {
		t.Fatalf("custom Elo: step %d elo %d", m.opp.step, m.opp.choice.elo)
	}
	do(t, m, keyMsg("down"))
	do(t, m, keyMsg("enter")) // Black
	if m.screen != screenMenu || !m.choice.computer || m.choice.elo != 1750 || m.choice.color != "black" {
		t.Fatalf("choice %+v", m.choice)
	}
	if last := rec.last(t); last.Opponent != "computer" || last.ComputerElo != 1750 || last.HumanColor != "black" {
		t.Fatalf("saved %+v", last)
	}
	if !strings.Contains(m.menuView(), "Computer (Elo 1750)") {
		t.Fatalf("menu:\n%s", m.menuView())
	}
}

func TestStrengthFor(t *testing.T) {
	stockfish := map[string]uci.Option{
		"uci_limitstrength": {Name: "UCI_LimitStrength", Type: "check"},
		"uci_elo":           {Name: "UCI_Elo", Type: "spin", Min: 1320, Max: 3190},
		"skill level":       {Name: "Skill Level", Type: "spin", Min: 0, Max: 20},
	}
	s := strengthFor(stockfish, 1600)
	if s.key() != strengthFor(stockfish, 1600).key() || len(s.options) != 3 || s.options[1] != [2]string{"UCI_Elo", "1600"} || s.depth != 0 {
		t.Fatalf("1600: %+v", s)
	}
	if s := strengthFor(stockfish, 9999); s.options[1] != [2]string{"UCI_Elo", "3190"} {
		t.Fatalf("Elo is clamped to the engine's range: %+v", s)
	}
	s = strengthFor(stockfish, 800)
	if s.options[0] != [2]string{"UCI_LimitStrength", "false"} || s.options[1] != [2]string{"Skill Level", "3"} || s.depth != 1 || s.think > time.Second {
		t.Fatalf("800: %+v", s)
	}
	s = strengthFor(stockfish, 0)
	if s.options[0] != [2]string{"UCI_LimitStrength", "false"} || s.options[1] != [2]string{"Skill Level", "20"} || s.think != 0 {
		t.Fatalf("max: %+v", s)
	}
	if s := strengthFor(nil, 1200); len(s.options) != 0 || s.depth != 3 {
		t.Fatalf("an engine without strength options gets a depth limit: %+v", s)
	}
}

func TestParseLevel(t *testing.T) {
	for in, want := range map[string]int{"medium": 1600, "MAX": 0, "1750": 1750, "Beginner": 800} {
		if got, ok := ParseLevel(in); !ok || got != want {
			t.Errorf("%q: %d %v", in, got, ok)
		}
	}
	for _, in := range []string{"", "godlike", "100", "9000"} {
		if _, ok := ParseLevel(in); ok {
			t.Errorf("%q should be rejected", in)
		}
	}
}

func TestEvalText(t *testing.T) {
	for e, want := range map[evalScore]string{
		{cp: 123}:             "+1.2",
		{cp: -40}:             "-0.4",
		{mate: 3}:             "M3",
		{mate: -2}:            "-M2",
		{over: true, mate: 1}: "1-0",
		{over: true}:          "0.0",
	} {
		if got := e.text(); got != want {
			t.Errorf("%+v: %q, want %q", e, got, want)
		}
	}
	if (evalScore{cp: 100}).whiteShare() <= 0.5 || (evalScore{mate: -1}).whiteShare() != 0 {
		t.Fatal("whiteShare")
	}
}

func TestEngineDisplayName(t *testing.T) {
	for path, want := range map[string]string{
		"/usr/games/stockfish":                     "Stockfish",
		"/opt/sf/stockfish-linux-x86-64-universal": "Stockfish",
		"/engines/komodo_14.exe":                   "Komodo",
		"":                                         "Computer",
	} {
		c := &engineCtl{path: path}
		if got := c.displayName(); got != want {
			t.Errorf("%q: %q, want %q", path, got, want)
		}
	}
}
