package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/chess/chess"
	"github.com/almeidazs/chess/clock"
	"github.com/almeidazs/chess/internal/config"
	"github.com/almeidazs/chess/internal/sound"
)

// Options configures the program.
type Options struct {
	// TimeControl, when set, skips the menu and starts a game right away.
	TimeControl *clock.TimeControl
	// Now is the time source (default time.Now). Overriding it makes
	// rendering deterministic, e.g. to record screenshots.
	Now func() time.Time

	// Config holds the saved preferences: theme, sound, last time control.
	Config config.Config
	// Theme overrides Config.Theme for this run only (see ThemeNames).
	Theme string
	// Mute turns the sound off for this run only.
	Mute bool
	// Sound plays the effects (default: silent).
	Sound sound.Player
	// Save persists the preferences whenever the player changes them
	// (default: nothing is saved). It runs outside the update loop.
	Save func(config.Config) error
}

// Model is the Bubble Tea model.
type Model struct {
	width, height int
	screen        screen
	menu          menuState
	startWith     *clock.TimeControl

	game        *chess.Game
	tc          clock.TimeControl
	clock       *clock.Clock // nil when playing without a clock
	orientation chess.Color  // side shown at the bottom; follows the side to move
	mode        mode

	cursorCol, cursorRow int  // keyboard cursor in screen cells (0,0 = top-left)
	showCursor           bool // only shown once the keyboard is used
	selected             chess.Square
	targets              []chess.Move // legal moves of the selected piece

	promo    []chess.Move // candidate promotions, Q R B N
	promoIdx int

	moveTimes []time.Duration // thinking time of each move, aligned with the history
	turnStart time.Time       // when the side to move started thinking

	drag    *dragState // piece being dragged with the mouse
	dropped bool       // the next move was made by dropping a dragged piece

	input     string  // move being typed after ':'
	inputErr  string  // why the last typed move was rejected
	lowWarned [2]bool // low-time sound already played, per color

	notice  string
	tickGen int

	help     bool   // help overlay open
	themeIdx int    // active theme, index into themes
	st       styles // styles derived from the theme
	cfg      config.Config
	saver    *saver
	player   sound.Player
	soundOn  bool

	animate bool       // post-move animations on the sprite board
	anim    *animation // running animation, nil when idle
	animGen int
	now     func() time.Time
}

// New creates the model.
func New(opts Options) *Model {
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	player := opts.Sound
	if player == nil {
		player = sound.Nop{}
	}
	m := &Model{
		startWith: opts.TimeControl,
		selected:  chess.NoSquare,
		animate:   true,
		now:       now,
		cfg:       opts.Config,
		saver:     &saver{fn: opts.Save},
		player:    player,
		soundOn:   opts.Config.Sound && !opts.Mute,
	}
	theme := opts.Config.Theme
	if opts.Theme != "" {
		theme = opts.Theme
	}
	idx, _ := themeIndex(theme) // unknown names fall back to the default
	m.setTheme(idx)
	m.menu.restore(opts.Config.TimeControl)
	return m
}

// Init starts the game right away when a time control was given.
func (m *Model) Init() tea.Cmd {
	if m.startWith != nil {
		return m.startGame(*m.startWith)
	}
	return nil
}

// Update routes Bubble Tea messages to their handlers.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tickMsg:
		return m, m.onTick(msg)
	case animMsg:
		return m, m.onAnimFrame(msg)
	case saveErrMsg:
		m.say("Não foi possível salvar as preferências")
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.anim = nil // any input skips the animation
		if m.help {
			return m, m.helpKey(msg.String())
		}
		if m.screen == screenMenu {
			return m, m.menuKey(msg)
		}
		return m, m.gameKey(msg.String())
	case tea.MouseMsg:
		return m, m.handleMouse(msg)
	}
	return m, nil
}
