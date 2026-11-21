package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/config"
	"github.com/almeidazs/tchess/internal/sound"
)

// recordPlayer remembers the effects it was asked to play.
type recordPlayer struct{ got []sound.Effect }

func (r *recordPlayer) Play(e sound.Effect) { r.got = append(r.got, e) }

// saveRecorder collects the configs the model asks to save.
type saveRecorder struct{ saved []config.Config }

func (s *saveRecorder) save(c config.Config) error {
	s.saved = append(s.saved, c)
	return nil
}

func (s *saveRecorder) last(t *testing.T) config.Config {
	t.Helper()
	if len(s.saved) == 0 {
		t.Fatal("nothing was saved")
	}
	return s.saved[len(s.saved)-1]
}

// press sends a key and runs the commands it returns (saves, batches),
// feeding their messages back, but not timers.
func press(m *Model, k string) {
	var msg tea.KeyMsg
	switch k {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	_, cmd := m.Update(msg)
	runCmd(m, cmd)
}

func runCmd(m *Model, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		switch msg := msg.(type) {
		case tea.BatchMsg:
			for _, c := range msg {
				runCmd(m, c)
			}
		case nil:
		default:
			m.Update(msg) // feed results back, as Bubble Tea would
		}
	case <-time.After(50 * time.Millisecond): // a timer: skip it
	}
}

func newPrefsModel(t *testing.T, opts Options) (*Model, *saveRecorder, *recordPlayer) {
	t.Helper()
	rec, player := &saveRecorder{}, &recordPlayer{}
	opts.Save, opts.Sound = rec.save, player
	m := New(opts)
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 50})
	m.animate = false
	return m, rec, player
}

func TestThemeCyclesAndSaves(t *testing.T) {
	m, rec, _ := newPrefsModel(t, Options{Config: config.Default()})
	if m.theme().Name != "Wood" {
		t.Fatalf("default theme %q", m.theme().Name)
	}
	for i := 1; i <= len(themes); i++ {
		press(m, "s")
		want := themes[i%len(themes)].Name
		if m.theme().Name != want || rec.last(t).Theme != want {
			t.Fatalf("press %d: theme %q saved %q, want %q", i, m.theme().Name, rec.last(t).Theme, want)
		}
	}
	if !strings.Contains(m.menuView(), "Theme: Wood") {
		t.Fatal("menu should confirm the theme change")
	}
}

func TestThemeChangesAccent(t *testing.T) {
	m, _, _ := newPrefsModel(t, Options{})
	before := m.st.accentColor
	press(m, "s")
	if m.st.accentColor == before {
		t.Fatal("accent should follow the theme")
	}
}

func TestThemeFromConfigAndOverride(t *testing.T) {
	m, _, _ := newPrefsModel(t, Options{Config: config.Config{Theme: "coral", Sound: true}})
	if m.theme().Name != "Coral" {
		t.Fatalf("theme from config: %q", m.theme().Name)
	}
	m, rec, _ := newPrefsModel(t, Options{Config: config.Config{Theme: "Coral", Sound: true}, Theme: "marble"})
	if m.theme().Name != "Marble" || m.cfg.Theme != "Coral" || len(rec.saved) != 0 {
		t.Fatalf("override: theme %q cfg %q saved %v", m.theme().Name, m.cfg.Theme, rec.saved)
	}
	if !HasTheme("OCEAN") || HasTheme("neon") {
		t.Fatal("HasTheme")
	}
}

func TestOldThemeNamesStillWork(t *testing.T) {
	for old, want := range map[string]string{"Madeira": "Wood", "grafite": "Graphite", "Mármore": "Marble", "marmore": "Marble"} {
		m, _, _ := newPrefsModel(t, Options{Config: config.Config{Theme: old}})
		if m.theme().Name != want {
			t.Errorf("config theme %q gave %q, want %q", old, m.theme().Name, want)
		}
	}
}

func TestSoundToggle(t *testing.T) {
	m, rec, player := newPrefsModel(t, Options{Config: config.Default()})
	press(m, "m")
	if m.soundOn || rec.last(t).Sound {
		t.Fatal("m should turn the sound off and save it")
	}
	press(m, "m")
	if !m.soundOn || !rec.last(t).Sound {
		t.Fatal("m should turn the sound back on")
	}
	if len(player.got) != 1 || player.got[0] != sound.Move {
		t.Fatalf("turning sound on should play a sample, got %v", player.got)
	}
}

func TestMuteIsNotSaved(t *testing.T) {
	m, rec, player := newPrefsModel(t, Options{Config: config.Default(), Mute: true})
	if m.soundOn {
		t.Fatal("--no-sound should start muted")
	}
	press(m, "s") // saving another preference must keep sound on in the file
	if !rec.last(t).Sound {
		t.Fatal("the mute flag leaked into the config")
	}
	press(m, "enter") // start a game: no sound should play
	click(m, chess.E2)
	click(m, chess.E4)
	if len(player.got) != 0 {
		t.Fatalf("muted game played %v", player.got)
	}
}

func TestSaveErrorIsShown(t *testing.T) {
	m := New(Options{Save: func(config.Config) error { return errors.New("disk full") }})
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 50})
	press(m, "s")
	if !strings.Contains(m.menu.notice, "save") {
		t.Fatalf("notice %q", m.menu.notice)
	}
}

func TestMenuRemembersTimeControl(t *testing.T) {
	m, rec, _ := newPrefsModel(t, Options{Config: config.Config{TimeControl: "5+3", Sound: true}})
	if m.menuEntries()[m.menu.cursor].tc != "5+3" {
		t.Fatalf("cursor on %v", m.menuEntries()[m.menu.cursor])
	}
	press(m, "down")
	press(m, "enter")
	if m.screen != screenGame || rec.last(t).TimeControl != "10+0" {
		t.Fatalf("screen %v saved %+v", m.screen, rec.saved)
	}

	m, _, _ = newPrefsModel(t, Options{Config: config.Config{TimeControl: "7+2"}})
	if m.menuEntries()[m.menu.cursor].kind != entryCustom {
		t.Fatalf("custom time control should select the custom item, got %v", m.menuEntries()[m.menu.cursor])
	}
	press(m, "enter")
	if !m.menu.editing || m.menu.input != "7+2" {
		t.Fatalf("custom input %q", m.menu.input)
	}

	m, rec, _ = newPrefsModel(t, Options{Config: config.Config{TimeControl: "none"}})
	if m.menuEntries()[m.menu.cursor].tc != "none" {
		t.Fatalf("cursor on %v", m.menuEntries()[m.menu.cursor])
	}
	press(m, "enter")
	if len(rec.saved) != 0 {
		t.Fatalf("unchanged time control was saved again: %v", rec.saved)
	}
}

func TestMoveSounds(t *testing.T) {
	tests := []struct {
		fen      string
		from, to chess.Square
		want     sound.Effect
	}{
		{chess.StartFEN, chess.E2, chess.E4, sound.Move},
		{"4k3/8/8/3p4/4P3/8/8/4K3 w - - 0 1", chess.E4, chess.D5, sound.Capture},
		{"4k3/8/8/8/8/8/8/R3K2R w KQ - 0 1", chess.E1, chess.G1, sound.Castle},
		{"4k3/8/4K3/8/8/8/8/R7 w - - 0 1", chess.A1, chess.A8, sound.GameEnd}, // mate
		{"4k3/8/8/8/8/8/8/3QK3 w - - 0 1", chess.D1, chess.D7, sound.Check},
	}
	for _, tt := range tests {
		m, _, player := newPrefsModel(t, Options{Config: config.Default()})
		m.startGame(clock.TimeControl{})
		g, err := chess.NewGameFromFEN(tt.fen)
		if err != nil {
			t.Fatal(err)
		}
		m.game = g
		click(m, tt.from)
		click(m, tt.to)
		if len(player.got) != 1 || player.got[0] != tt.want {
			t.Errorf("%s %v-%v: sounds %v, want %v", tt.fen, tt.from, tt.to, player.got, tt.want)
		}
	}
}

func TestPromotionSound(t *testing.T) {
	m, _, player := newPrefsModel(t, Options{Config: config.Default()})
	m.startGame(clock.TimeControl{})
	g, _ := chess.NewGameFromFEN("1k6/7P/p7/8/8/8/8/4K3 w - - 0 1") // the pawn keeps mating material
	m.game = g
	click(m, chess.H7)
	click(m, chess.H8)
	key(m, "n") // knight: no check, so the promotion sound plays
	if len(player.got) != 1 || player.got[0] != sound.Promote {
		t.Fatalf("sounds %v", player.got)
	}
}

func TestLowTimeWarnsOnce(t *testing.T) {
	now := time.Unix(0, 0)
	player := &recordPlayer{}
	m := New(Options{Now: func() time.Time { return now }, Sound: player, Config: config.Default()})
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 50})
	m.animate = false
	m.startGame(clock.TimeControl{Base: 15 * time.Second})
	click(m, chess.E2) // White's first move starts Black's clock
	click(m, chess.E4)
	player.got = nil
	now = now.Add(4 * time.Second)
	m.Update(tickMsg{m.tickGen})
	if len(player.got) != 0 {
		t.Fatalf("warned too early: %v", player.got)
	}
	now = now.Add(2 * time.Second) // 9 s left
	m.Update(tickMsg{m.tickGen})
	m.Update(tickMsg{m.tickGen})
	if len(player.got) != 1 || player.got[0] != sound.LowTime {
		t.Fatalf("sounds %v, want one low-time warning", player.got)
	}
}

func TestHelpOverlay(t *testing.T) {
	m := newGame(t, 140, 50, "5+3")
	key(m, "?")
	if !m.help || !strings.Contains(m.View(), "Help") {
		t.Fatal("? should open the help")
	}
	key(m, "d") // other shortcuts are inert while the help is open
	if m.mode != modePlay {
		t.Fatalf("mode %v", m.mode)
	}
	key(m, "s") // theme still works, and the card shows it
	if !strings.Contains(m.View(), "theme: Forest") {
		t.Fatal("help should show the current theme")
	}
	key(m, "?")
	if m.help {
		t.Fatal("? should close the help")
	}

	m = New(Options{})
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	key(m, "?")
	if !strings.Contains(m.View(), "play or open") {
		t.Fatal("menu help should list the menu keys")
	}
	key(m, "esc")
	if m.help || m.screen != screenMenu {
		t.Fatal("esc should close the menu help without leaving")
	}
}

func TestHelpFitsSmallestBoard(t *testing.T) {
	m := newGame(t, 58, 14, "none")
	key(m, "?")
	box := m.helpBox(14)
	if h := strings.Count(box, "\n") + 1; h > 14 {
		t.Fatalf("help is %d rows tall on a 14-row terminal", h)
	}
	if w := lipgloss.Width(box); w > 58 {
		t.Fatalf("help is %d columns wide on a 58-column terminal", w)
	}
	if lines := strings.Split(m.View(), "\n"); len(lines) > 14 {
		t.Fatalf("view grew to %d rows", len(lines))
	}
}
