package tui

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/tchess/chess"
)

// The PGN picker lists the games folder, newest first, followed by an
// entry to type any path. A file with several games opens a list of them.
// A finished game opens on its final position; an unfinished one asks how
// to time the rest of it (see clocks.go).

const (
	openInnerW   = 44 // content width
	openVisible  = 8  // rows of files or games shown at once
	openTypePath = "Digitar caminho…"
)

// openPicker shows the PGN picker.
func (m *Model) openPicker() {
	files, err := m.files.listGames()
	m.open = openState{files: files}
	if err != nil {
		m.open.err = "Não foi possível ler a pasta de partidas"
	}
	m.screen = screenOpen
}

// showGames lists the games of a file with several of them.
func (m *Model) showGames(games []*chess.PGNGame, source string, fromFiles bool) {
	m.open.games, m.open.source, m.open.gameCursor, m.open.fromFiles = games, source, 0, fromFiles
	m.open.editing = false
	m.screen = screenOpen
}

// openKey handles a key on the PGN picker.
func (m *Model) openKey(key string) tea.Cmd {
	o := &m.open
	switch {
	case o.games != nil:
		return m.gamesKey(key)
	case o.editing:
		return m.pathKey(key)
	}
	n := len(o.files) + 1
	switch key {
	case "up", "k":
		o.cursor = (o.cursor + n - 1) % n
	case "down", "j", "tab":
		o.cursor = (o.cursor + 1) % n
	case "enter", " ", "space":
		return m.chooseOpen()
	case "esc", "q":
		m.screen = screenMenu
	}
	return nil
}

// pathKey handles typing a path.
func (m *Model) pathKey(key string) tea.Cmd {
	o := &m.open
	switch key {
	case "esc":
		o.editing, o.input = false, ""
	case "enter":
		return m.openFile(expandHome(strings.TrimSpace(o.input)))
	case "backspace":
		if _, size := utf8.DecodeLastRuneInString(o.input); size > 0 {
			o.input = o.input[:len(o.input)-size]
		}
	case "space":
		o.input += " "
	default:
		if utf8.RuneCountInString(key) == 1 {
			o.input += key
		}
	}
	o.err = ""
	return nil
}

// gamesKey handles the list of a file's games.
func (m *Model) gamesKey(key string) tea.Cmd {
	o := &m.open
	n := len(o.games)
	switch key {
	case "up", "k":
		o.gameCursor = (o.gameCursor + n - 1) % n
	case "down", "j", "tab":
		o.gameCursor = (o.gameCursor + 1) % n
	case "enter", " ", "space":
		return m.openRecord(o.games[o.gameCursor], o.source, screenOpen)
	case "esc", "q":
		o.games, o.err = nil, ""
		if !o.fromFiles {
			m.screen = screenMenu
		}
	}
	return nil
}

// chooseOpen opens the highlighted file, or starts typing a path.
func (m *Model) chooseOpen() tea.Cmd {
	o := &m.open
	if o.cursor < len(o.files) {
		return m.openFile(o.files[o.cursor].path)
	}
	o.editing, o.input, o.err = true, "", ""
	return nil
}

// openFile reads a PGN file: its only game opens directly, several are
// listed.
func (m *Model) openFile(path string) tea.Cmd {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		m.open.err = "Arquivo não encontrado"
		return nil
	case errors.Is(err, fs.ErrPermission):
		m.open.err = "Sem permissão para ler o arquivo"
		return nil
	case err != nil:
		m.open.err = "Não foi possível ler o arquivo"
		return nil
	}
	games, perr := chess.ParsePGNGames(string(data))
	if len(games) == 0 {
		m.open.err = "PGN inválido ou com lance ilegal"
		return nil
	}
	m.open.editing, m.open.err = false, ""
	if len(games) == 1 && perr == nil {
		return m.openRecord(games[0], path, screenOpen)
	}
	m.showGames(games, path, true)
	if skipped := countErrors(perr); skipped > 0 {
		m.open.err = fmt.Sprintf("%d partida(s) inválida(s) ignorada(s)", skipped)
	}
	return nil
}

// countErrors counts the errors joined in err.
func countErrors(err error) int {
	if err == nil {
		return 0
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		return len(joined.Unwrap())
	}
	return 1
}

// openRecord opens one game. A finished game shows its final position; an
// unfinished one first asks how to time the rest. back is where Esc on
// that question returns.
func (m *Model) openRecord(pg *chess.PGNGame, source string, back screen) tea.Cmd {
	if pg.Game.IsOver() || pg.Result != chess.Ongoing {
		cmd := m.loadRecord(pg, clocksOff, clockNoTime)
		m.savedPath = source // already on disk: e should not make a copy
		return cmd
	}
	m.askClocks(pg, source, back)
	return nil
}

// gameLabel names a game in the list: number and players.
func gameLabel(i int, pg *chess.PGNGame) string {
	player := func(tag string, c chess.Color) string {
		if v := pg.Tag(tag); v != "" && v != "?" {
			return v
		}
		return colorName(c)
	}
	return fmt.Sprintf("%d. %s x %s", i+1, player("White", chess.White), player("Black", chess.Black))
}

// gameDetail sums a game up: result and length.
func gameDetail(pg *chess.PGNGame) string {
	result := outcomeScore(pg.Result)
	if pg.Game.IsOver() {
		result = outcomeScore(pg.Game.Outcome())
	}
	return fmt.Sprintf("%s · %d lances", result, (pg.Game.Ply()+1)/2)
}

// window is the first row shown of n rows so that cursor stays visible;
// past the end (the "type a path" entry) the list shows its last rows.
func window(cursor, n int) int {
	if cursor >= n {
		return max(n-openVisible, 0)
	}
	return max(cursor-openVisible+1, 0)
}

// openLines returns the picker's content and, for each visible entry, its
// line and entry index.
func (m *Model) openLines() (lines []string, rows map[int]int) {
	o := &m.open
	rows = map[int]int{}
	lines = []string{center(m.st.title.Render("Abrir PGN"), openInnerW)}

	entry := func(i, cursor int, label, right string) {
		rows[len(lines)] = i
		width := openInnerW - 4
		label = truncate(label, width-lipgloss.Width(right)-2)
		row := pad(label, width-lipgloss.Width(right)) + right
		if i == cursor {
			lines = append(lines, m.st.accent.Render("▸ "+row+"  "))
		} else {
			lines = append(lines, textStyle.Render("  "+row+"  "))
		}
	}

	if o.games != nil {
		sub := fmt.Sprintf("%s · %d partidas", filepath.Base(o.source), len(o.games))
		lines = append(lines, center(dimStyle.Render(truncate(sub, openInnerW)), openInnerW), "")
		start := window(o.gameCursor, len(o.games))
		for i := start; i < len(o.games) && i < start+openVisible; i++ {
			entry(i, o.gameCursor, gameLabel(i, o.games[i]), gameDetail(o.games[i]))
		}
	} else {
		if m.files.gamesDir != "" {
			lines = append(lines, center(dimStyle.Render(truncate(tildePath(m.files.gamesDir), openInnerW)), openInnerW))
		}
		lines = append(lines, "")
		if len(o.files) == 0 {
			lines = append(lines, dimStyle.Render("  Nenhuma partida ainda: as terminadas"), dimStyle.Render("  aparecem aqui"))
		}
		start := window(o.cursor, len(o.files))
		for i := start; i < len(o.files) && i < start+openVisible; i++ {
			f := o.files[i]
			entry(i, o.cursor, strings.TrimSuffix(f.name, ".pgn"), dimStyle.Render(f.modTime.Format("02/01 15:04")))
		}
		lines = append(lines, "")
		entry(len(o.files), o.cursor, openTypePath, "")
	}

	lines = append(lines, "")
	switch {
	case o.editing:
		lines = append(lines, textStyle.Render("caminho: ")+m.st.accent.Render(truncate(o.input, openInnerW-10)+"▌"))
	case o.err != "":
		lines = append(lines, dangerStyle.Render(truncate(o.err, openInnerW)))
	default:
		lines = append(lines, "")
	}
	help := [][2]string{{"↑↓", "escolher"}, {"enter", "abrir"}, {"esc", "voltar"}}
	return append(lines, "", m.helpLine(help)), rows
}

// openView renders the PGN picker.
func (m *Model) openView() string {
	lines, _ := m.openLines()
	return m.st.menuBox.Width(openInnerW + 2*menuPadX).Render(strings.Join(lines, "\n"))
}

// openClick acts on the entry under the pointer.
func (m *Model) openClick(x, y int) tea.Cmd {
	if m.open.editing {
		return nil
	}
	view := m.openView()
	ox, oy := centerOffset(m.width, m.height, lipgloss.Width(view), lipgloss.Height(view))
	lines, rows := m.openLines()
	row := y - (oy + 1 + menuPadY)
	col := x - (ox + 1 + menuPadX)
	if row < 0 || row >= len(lines) || col < 0 || col >= openInnerW {
		return nil
	}
	i, ok := rows[row]
	if !ok {
		return nil
	}
	if m.open.games != nil {
		m.open.gameCursor = i
		return m.openRecord(m.open.games[i], m.open.source, screenOpen)
	}
	m.open.cursor = i
	return m.chooseOpen()
}
