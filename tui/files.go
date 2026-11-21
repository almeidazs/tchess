package tui

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/internal/atomicfile"
)

// serial runs background writes in order: each gets a sequence number when
// it is scheduled, and a write that finishes after a newer one is dropped,
// so the file always ends up with the latest content.
type serial struct {
	mu   sync.Mutex
	seq  int // last number handed out (update loop only)
	done int // last number written
}

// saveErrMsg reports a background write that failed; what is shown to
// the player.
type saveErrMsg struct{ what string }

// run returns a command performing fn in order; on failure it reports what.
func (s *serial) run(fn func() error, what string) tea.Cmd {
	s.seq++
	seq := s.seq
	return func() tea.Msg {
		s.mu.Lock()
		defer s.mu.Unlock()
		if seq < s.done {
			return nil
		}
		s.done = seq
		if err := fn(); err != nil {
			return saveErrMsg{what}
		}
		return nil
	}
}

// fileStore reads and writes the game files: the unfinished game kept for
// resuming, exported PGNs and the list the PGN picker shows. An empty path
// disables the matching feature.
type fileStore struct {
	gamesDir   string
	resumePath string
	resumeW    serial
}

// saveResume keeps the unfinished game.
func (f *fileStore) saveResume(pgn string) tea.Cmd {
	if f.resumePath == "" {
		return nil
	}
	path := f.resumePath
	return f.resumeW.run(func() error { return atomicfile.Write(path, []byte(pgn)) },
		"Não foi possível salvar a partida para retomar")
}

// clearResume forgets the unfinished game.
func (f *fileStore) clearResume() tea.Cmd {
	if f.resumePath == "" {
		return nil
	}
	path := f.resumePath
	return f.resumeW.run(func() error {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}, "Não foi possível apagar a partida salva")
}

// loadResume reads the unfinished game, if there is one. A file that is
// unreadable, finished or has no moves is ignored.
func (f *fileStore) loadResume() *savedGame {
	if f.resumePath == "" {
		return nil
	}
	pg, err := readPGNFile(f.resumePath)
	if err != nil || pg.Game.IsOver() || pg.Result != chess.Ongoing || pg.Game.Ply() == 0 {
		return nil
	}
	tc := "sem relógio"
	if t, ok := parsePGNTimeControl(pg.Tag("TimeControl")); ok && !t.Unlimited() {
		tc = t.String()
	}
	pos := pg.Game.Position()
	return &savedGame{record: pg, label: fmt.Sprintf("%s · lance %d", tc, pos.FullmoveNumber())}
}

// exportedMsg reports an export: the file written, or the error.
type exportedMsg struct {
	path string
	err  error
}

// export writes pgn to a new file named after the moment of the export.
func (f *fileStore) export(pgn, name string) tea.Cmd {
	if f.gamesDir == "" {
		return func() tea.Msg { return exportedMsg{err: errors.New("sem pasta de partidas")} }
	}
	path := filepath.Join(f.gamesDir, name+".pgn")
	return func() tea.Msg {
		return exportedMsg{path: path, err: atomicfile.Write(path, []byte(pgn))}
	}
}

// listGames returns the PGN files of the games folder, newest first.
func (f *fileStore) listGames() ([]pgnFile, error) {
	if f.gamesDir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(f.gamesDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []pgnFile
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".pgn") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, pgnFile{name: e.Name(), path: filepath.Join(f.gamesDir, e.Name()), modTime: info.ModTime()})
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].modTime.Equal(out[j].modTime) {
			return out[i].modTime.After(out[j].modTime)
		}
		return out[i].name > out[j].name // exported names are timestamps
	})
	return out, nil
}

// readPGNFile reads and parses a PGN file.
func readPGNFile(path string) (*chess.PGNGame, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return chess.ParsePGN(string(data))
}

// expandHome replaces a leading ~ with the home directory.
func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}

// tildePath shortens a path under the home directory to ~/...
func tildePath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if rel, err := filepath.Rel(home, path); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.Join("~", rel)
	}
	return path
}
