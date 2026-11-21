package tui

import (
	"errors"
	"io/fs"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The PGN picker lists the games folder, newest first, followed by an
// entry to type any path. A finished game opens on its result, ready to be
// reviewed; an unfinished one can be played on, without a clock.

const (
	openInnerW   = 44 // content width
	openVisible  = 8  // file rows shown at once
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

// openKey handles a key on the PGN picker.
func (m *Model) openKey(key string) tea.Cmd {
	o := &m.open
	if o.editing {
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

// chooseOpen opens the highlighted file, or starts typing a path.
func (m *Model) chooseOpen() tea.Cmd {
	o := &m.open
	if o.cursor < len(o.files) {
		return m.openFile(o.files[o.cursor].path)
	}
	o.editing, o.input, o.err = true, "", ""
	return nil
}

// openFile loads a PGN file into a new game.
func (m *Model) openFile(path string) tea.Cmd {
	if path == "" {
		return nil
	}
	pg, err := readPGNFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		m.open.err = "Arquivo não encontrado"
		return nil
	case err != nil && errors.Is(err, fs.ErrPermission):
		m.open.err = "Sem permissão para ler o arquivo"
		return nil
	case err != nil:
		m.open.err = "PGN inválido ou com lance ilegal"
		return nil
	}
	m.open.editing = false
	cmd := m.loadRecord(pg, false)
	if m.mode == modeGameOver {
		m.savedPath = path // already on disk: e should not make a copy
	}
	return cmd
}

// openWindow is the first file row shown, keeping the cursor visible.
// On the "type a path" entry the list shows its end, which sits right
// above it.
func (m *Model) openWindow() int {
	o := &m.open
	if o.cursor >= len(o.files) {
		return max(len(o.files)-openVisible, 0)
	}
	return max(o.cursor-openVisible+1, 0)
}

// openLines returns the picker's content and, for each visible entry, its
// line and entry index.
func (m *Model) openLines() (lines []string, rows map[int]int) {
	o := &m.open
	rows = map[int]int{}
	lines = []string{center(m.st.title.Render("Abrir PGN"), openInnerW)}
	if m.files.gamesDir != "" {
		lines = append(lines, center(dimStyle.Render(truncate(tildePath(m.files.gamesDir), openInnerW)), openInnerW))
	}
	lines = append(lines, "")

	entry := func(i int, label, right string) {
		rows[len(lines)] = i
		label = truncate(label, openInnerW-4-lipgloss.Width(right)-1)
		row := pad(label, openInnerW-4-lipgloss.Width(right)) + right
		if i == o.cursor {
			lines = append(lines, m.st.accent.Render("▸ "+row+"  "))
		} else {
			lines = append(lines, textStyle.Render("  ")+textStyle.Render(row)+"  ")
		}
	}
	if len(o.files) == 0 {
		lines = append(lines, dimStyle.Render("  Nenhuma partida ainda: as terminadas"), dimStyle.Render("  aparecem aqui"))
	}
	start := m.openWindow()
	for i := start; i < len(o.files) && i < start+openVisible; i++ {
		f := o.files[i]
		entry(i, strings.TrimSuffix(f.name, ".pgn"), dimStyle.Render(f.modTime.Format("02/01 15:04")))
	}
	lines = append(lines, "")
	entry(len(o.files), openTypePath, "")

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
	if i, ok := rows[row]; ok {
		m.open.cursor = i
		return m.chooseOpen()
	}
	return nil
}
