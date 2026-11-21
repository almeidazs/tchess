package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/internal/config"
)

// selectEntry moves the menu cursor to the first entry of the given kind.
func selectEntry(t *testing.T, m *Model, kind entryKind) {
	t.Helper()
	for i, e := range m.menuEntries() {
		if e.kind == kind {
			m.menu.cursor = i
			return
		}
	}
	t.Fatalf("no menu entry of kind %v", kind)
}

func TestMenuListsOptions(t *testing.T) {
	m, _, _ := newPrefsModel(t, Options{Config: config.Default()})
	view := m.menuView()
	for _, want := range []string{"Opções", "Tema", "Madeira ›", "Girar tabuleiro", "sim", "Abrir PGN"} {
		if !strings.Contains(view, want) {
			t.Fatalf("menu lacks %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Continuar") {
		t.Fatal("no saved game, no Continuar entry")
	}
}

func TestMenuFlipEntry(t *testing.T) {
	m, rec, _ := newPrefsModel(t, Options{Config: config.Default()})
	selectEntry(t, m, entryFlip)
	press(m, "enter")
	if m.flip || rec.last(t).Flip || !strings.Contains(m.menuView(), "não") {
		t.Fatal("the menu entry should turn board turning off and save it")
	}

	// Clicking the entry toggles it back.
	view := m.menuView()
	ox, oy := centerOffset(m.width, m.height, lipgloss.Width(view), lipgloss.Height(view))
	_, entryLine := m.menuLines()
	selectEntry(t, m, entryFlip)
	m.Update(tea.MouseMsg{X: ox + 1 + menuPadX + 2, Y: oy + 1 + menuPadY + entryLine[m.menu.cursor],
		Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if !m.flip {
		t.Fatal("clicking the entry should toggle it")
	}
}

func TestFlipOffKeepsWhiteBelow(t *testing.T) {
	cfg := config.Default()
	cfg.Flip = false
	m, rec, _ := newPrefsModel(t, Options{Config: cfg})
	m.startGame(clockNone())
	click(m, chess.E2)
	click(m, chess.E4)
	if m.orientation != chess.White {
		t.Fatal("with turning off the board stays on White's side")
	}
	press(m, "f")
	if m.orientation != chess.Black || !rec.last(t).Flip {
		t.Fatal("f turns the board to the side to move at once and saves it")
	}
	if !strings.Contains(m.statusView(), "Girar tabuleiro: sim") {
		t.Fatalf("status %q", m.statusView())
	}
}

func TestThemePicker(t *testing.T) {
	m, rec, _ := newPrefsModel(t, Options{Config: config.Default()})
	selectEntry(t, m, entryTheme)
	press(m, "enter")
	if m.screen != screenThemes {
		t.Fatalf("screen %v", m.screen)
	}
	view := m.View()
	for _, th := range themes {
		if !strings.Contains(view, th.Name) {
			t.Fatalf("picker lacks %s", th.Name)
		}
	}
	press(m, "down")
	if m.theme().Name != "Floresta" {
		t.Fatal("moving the cursor previews the theme")
	}
	press(m, "esc")
	if m.screen != screenMenu || m.theme().Name != "Madeira" || len(rec.saved) != 0 {
		t.Fatal("esc restores the previous theme without saving")
	}

	press(m, "enter")
	press(m, "down")
	press(m, "down")
	press(m, "enter")
	if m.screen != screenMenu || m.theme().Name != "Oceano" || rec.last(t).Theme != "Oceano" {
		t.Fatalf("enter keeps and saves the theme: %q", m.theme().Name)
	}
	if !strings.Contains(m.menuView(), "Oceano ›") {
		t.Fatal("the menu shows the chosen theme")
	}
}

func TestThemePreviewSizes(t *testing.T) {
	for _, size := range [][2]int{{160, 50}, {100, 40}, {80, 30}, {50, 20}} {
		m, _, _ := newPrefsModel(t, Options{Config: config.Default()})
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		m.openThemes()
		view := m.themesView()
		if w, h := lipgloss.Width(view), lipgloss.Height(view); (w > size[0] || h > size[1]) && size[0] >= 50 {
			t.Errorf("%v: picker is %dx%d", size, w, h)
		}
	}
	// The preview uses the real board renderer.
	m, _, _ := newPrefsModel(t, Options{Config: config.Default()})
	l := previewLayout(160, 50)
	if !l.pixel || len(m.themePreview(0, l)) != 8*l.cellH {
		t.Fatal("large terminals get a sprite preview")
	}
}

func TestThemePickerClick(t *testing.T) {
	m, _, _ := newPrefsModel(t, Options{Config: config.Default()})
	m.openThemes()
	view := m.themesView()
	ox, oy := centerOffset(m.width, m.height, lipgloss.Width(view), lipgloss.Height(view))
	at := func(i int) {
		m.Update(tea.MouseMsg{X: ox + 1 + menuPadX + 2, Y: oy + 1 + menuPadY + 2 + i,
			Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	}
	at(4)
	if m.theme().Name != "Coral" || m.screen != screenThemes {
		t.Fatal("a click previews")
	}
	at(4)
	if m.screen != screenMenu || m.cfg.Theme != "Coral" {
		t.Fatalf("a second click keeps it: screen %v theme %q", m.screen, m.cfg.Theme)
	}
}

func writePGN(t *testing.T, dir, name, text string, mod time.Time) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mod, mod); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOpenPicker(t *testing.T) {
	dir := t.TempDir()
	games := filepath.Join(dir, "games")
	writePGN(t, games, "old.pgn", "1. e4 e5 *", time.Unix(1000, 0))
	writePGN(t, games, "new.pgn", "1. d4 d5 2. c4 *", time.Unix(2000, 0))
	writePGN(t, games, "notes.txt", "not a game", time.Unix(3000, 0))
	clk := &fakeTime{time.Unix(5000, 0)}
	m := newDiskModel(t, dir, clk, config.Default())

	selectEntry(t, m, entryOpen)
	press(m, "enter")
	if m.screen != screenOpen || len(m.open.files) != 2 || m.open.files[0].name != "new.pgn" {
		t.Fatalf("files %v", m.open.files)
	}
	if strings.Contains(m.View(), "notes") {
		t.Fatal("only .pgn files are listed")
	}
	press(m, "enter")
	if m.screen != screenGame || m.game.Ply() != 3 || m.clock != nil {
		t.Fatalf("opened game: screen %v ply %d", m.screen, m.game.Ply())
	}
}

func TestOpenPickerTypedPath(t *testing.T) {
	dir := t.TempDir()
	other := writePGN(t, filepath.Join(dir, "elsewhere"), "game.pgn", "[Result \"0-1\"]\n\n1. f3 e5 2. g4 Qh4# 0-1", time.Unix(0, 0))
	m := newDiskModel(t, dir, &fakeTime{time.Unix(0, 0)}, config.Default())
	m.openPicker()
	if !strings.Contains(m.View(), "Nenhuma partida ainda") {
		t.Fatal("an empty folder says so")
	}
	press(m, "enter") // the only entry: type a path
	if !m.open.editing {
		t.Fatal("enter on the last entry starts typing")
	}
	for _, r := range "missing.pgn" {
		press(m, string(r))
	}
	press(m, "enter")
	if m.screen != screenOpen || m.open.err != "Arquivo não encontrado" {
		t.Fatalf("missing file: %q", m.open.err)
	}
	m.open.editing, m.open.input = true, other
	press(m, "enter")
	if m.screen != screenGame || m.mode != modeGameOver || m.game.Termination() != chess.Checkmate {
		t.Fatalf("screen %v mode %v", m.screen, m.mode)
	}
}

func TestOpenPickerBadFile(t *testing.T) {
	dir := t.TempDir()
	writePGN(t, filepath.Join(dir, "games"), "bad.pgn", "1. e4 e4 *", time.Unix(0, 0))
	m := newDiskModel(t, dir, &fakeTime{time.Unix(0, 0)}, config.Default())
	m.openPicker()
	press(m, "enter")
	if m.screen != screenOpen || !strings.Contains(m.open.err, "PGN inválido") {
		t.Fatalf("bad file: screen %v err %q", m.screen, m.open.err)
	}
	press(m, "esc")
	if m.screen != screenMenu {
		t.Fatal("esc goes back to the menu")
	}
}
