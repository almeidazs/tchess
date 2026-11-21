package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestMenuClickStartsGame(t *testing.T) {
	m := New(Options{})
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	for y, line := range lines {
		if x := strings.Index(line, "Blitz"); x >= 0 && strings.Contains(line, "5+3") {
			m.Update(tea.MouseMsg{X: len([]rune(line[:x])), Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
			break
		}
	}
	if m.screen != screenGame || m.tc.String() != "5+3" {
		t.Fatalf("menu click: screen=%v tc=%v", m.screen, m.tc)
	}
}

func TestMenuCustomTimeControl(t *testing.T) {
	m := New(Options{})
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	for range len(menuItems) - 1 {
		key(m, "down")
	}
	key(m, "enter")
	for _, r := range "7+2" {
		key(m, string(r))
	}
	key(m, "enter")
	if m.screen != screenGame || m.tc.String() != "7+2" {
		t.Fatalf("custom: screen=%v tc=%v err=%q", m.screen, m.tc, m.menu.err)
	}
}
