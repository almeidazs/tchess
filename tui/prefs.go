package tui

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/chess/chess"
	"github.com/almeidazs/chess/clock"
	"github.com/almeidazs/chess/internal/config"
	"github.com/almeidazs/chess/internal/sound"
)

// lowTime is when a clock turns red and the warning sound plays.
const lowTime = 10 * time.Second

// saveErrMsg reports that the preferences could not be written.
type saveErrMsg struct{ err error }

// saver writes the preferences in the background. Saves may finish out of
// order, so each carries a sequence number and older ones are dropped.
type saver struct {
	fn func(config.Config) error

	mu   sync.Mutex
	seq  int // last sequence number handed out (update loop only)
	done int // last sequence number written
}

// save returns a command that writes cfg, or nil when saving is disabled.
func (s *saver) save(cfg config.Config) tea.Cmd {
	if s.fn == nil {
		return nil
	}
	s.seq++
	seq := s.seq
	return func() tea.Msg {
		s.mu.Lock()
		defer s.mu.Unlock()
		if seq < s.done {
			return nil
		}
		s.done = seq
		if err := s.fn(cfg); err != nil {
			return saveErrMsg{err}
		}
		return nil
	}
}

// prefKey handles the keys that work on every screen: help, theme and
// sound. It reports whether the key was one of them.
func (m *Model) prefKey(key string) (tea.Cmd, bool) {
	switch key {
	case "?":
		m.help = true
		return nil, true
	case "s":
		return m.cycleTheme(), true
	case "m":
		return m.toggleSound(), true
	}
	return nil, false
}

// helpKey handles a key while the help overlay is open.
func (m *Model) helpKey(key string) tea.Cmd {
	switch key {
	case "?", "esc", "enter", "q":
		m.help = false
	case "s":
		return m.cycleTheme()
	case "m":
		return m.toggleSound()
	}
	return nil
}

// cycleTheme switches to the next theme and saves the choice.
func (m *Model) cycleTheme() tea.Cmd {
	m.setTheme((m.themeIdx + 1) % len(themes))
	m.cfg.Theme = m.theme().Name
	m.say("Tema: " + m.theme().Name)
	return m.saver.save(m.cfg)
}

// toggleSound turns the sound on or off and saves the choice.
func (m *Model) toggleSound() tea.Cmd {
	m.soundOn = !m.soundOn
	m.cfg.Sound = m.soundOn
	if m.soundOn {
		m.say("Som ligado")
		m.player.Play(sound.Move) // let the player hear it works
	} else {
		m.say("Som desligado")
	}
	return m.saver.save(m.cfg)
}

// rememberTimeControl saves the time control chosen in the menu.
func (m *Model) rememberTimeControl(tc clock.TimeControl) tea.Cmd {
	s := timeControlKey(tc)
	if s == m.cfg.TimeControl {
		return nil
	}
	m.cfg.TimeControl = s
	return m.saver.save(m.cfg)
}

// timeControlKey is how a time control is stored in the config file.
func timeControlKey(tc clock.TimeControl) string {
	if tc.Unlimited() {
		return "none"
	}
	return tc.String()
}

// say shows a short message on the current screen.
func (m *Model) say(msg string) {
	if m.screen == screenMenu {
		m.menu.notice = msg
		return
	}
	m.notice = msg
}

// sfx plays an effect when the sound is on.
func (m *Model) sfx(e sound.Effect) {
	if m.soundOn {
		m.player.Play(e)
	}
}

// moveEffect picks the sound of a move just played; inCheck tells whether
// it gave check.
func moveEffect(mv chess.Move, inCheck bool) sound.Effect {
	switch {
	case inCheck:
		return sound.Check
	case mv.IsPromotion():
		return sound.Promote
	case mv.IsCastle():
		return sound.Castle
	case mv.IsCapture():
		return sound.Capture
	}
	return sound.Move
}

// checkLowTime plays the warning once per game and side, when the side to
// move drops below lowTime. Games that start below it never warn.
func (m *Model) checkLowTime() {
	if m.clock == nil || m.tc.Base <= lowTime {
		return
	}
	c := m.game.Turn()
	if !m.lowWarned[c] && m.clock.Remaining(c) < lowTime {
		m.lowWarned[c] = true
		m.sfx(sound.LowTime)
	}
}
