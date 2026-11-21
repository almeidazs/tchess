package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/tchess/chess"
	"github.com/almeidazs/tchess/clock"
	"github.com/almeidazs/tchess/internal/config"
	"github.com/almeidazs/tchess/internal/sound"
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

	// GamesDir is where exported games are written and where the PGN
	// picker looks for files (default: exporting is disabled).
	GamesDir string
	// ResumePath is the file that keeps the unfinished game, so it can be
	// continued after the program closes (default: games are not kept).
	ResumePath string
	// Import, when set, opens these games (read from ImportPath) instead
	// of the menu: one opens directly, several are listed.
	Import     []*chess.PGNGame
	ImportPath string

	// EnginePath is the UCI engine to play against and analyze with
	// (default: none, which disables both).
	EnginePath string
	// Computer, when set, overrides the saved opponent for this run: the
	// computer at this strength (a level name or an Elo, see ParseLevel).
	Computer string
	// Color overrides the human's side against the computer for this run:
	// "white", "black" or "random".
	Color string
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
	orientation chess.Color  // side shown at the bottom
	mode        mode
	flip        bool // the board turns to face the side to move

	reviewing bool // showing an earlier position instead of the live one
	reviewPly int  // while reviewing: plies played in the position shown

	moveClocks []moveClock // clock around each move (clock games), from ply clockFrom on
	clockFrom  int         // first ply with a clock: later than 0 when a clock joined midway

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

	themes themePicker
	open   openState
	clocks clockChoice
	opp    opponentPicker
	resume *savedGame // unfinished game offered by the menu

	savedPath    string // file the finished game was saved to, "" until then
	resultHidden bool   // the result card is put aside to see the board
	undoTarget   int    // while a take-back is asked: plies the game would keep

	engine     *engineCtl
	engineName string // the engine's own name, once it has started
	engineGame int    // game number, so the engine knows a new game started
	choice     opponentChoice
	vs         opponent // the opponent of the current game
	thinking   bool     // the computer is searching for its move

	evals       map[int]evalScore // evaluations by ply, for review and the end
	evalPending int               // ply being evaluated, -1 for none

	hint           chess.Move // the engine's suggestion for the position at hintPly
	hintPly        int
	hintSet        bool
	assistTakeback bool // the human took moves back against the computer
	assistHint     bool // the human asked for hints
	files          *fileStore

	help     bool   // help overlay open
	themeIdx int    // active theme, index into themes
	st       styles // styles derived from the theme
	cfg      config.Config
	saver    *saver
	player   sound.Player
	soundOn  bool

	importPGN  []*chess.PGNGame // games to open at startup
	importPath string

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
		startWith:   opts.TimeControl,
		selected:    chess.NoSquare,
		animate:     true,
		now:         now,
		cfg:         opts.Config,
		saver:       &saver{fn: opts.Save},
		player:      player,
		soundOn:     opts.Config.Sound && !opts.Mute,
		flip:        opts.Config.Flip,
		files:       &fileStore{gamesDir: opts.GamesDir, resumePath: opts.ResumePath},
		importPGN:   opts.Import,
		importPath:  opts.ImportPath,
		engine:      &engineCtl{path: opts.EnginePath},
		evals:       map[int]evalScore{},
		evalPending: -1,
		choice: opponentChoice{
			computer: opts.Config.Opponent == "computer",
			elo:      opts.Config.ComputerElo,
			color:    opts.Config.HumanColor,
		},
	}
	if elo, ok := ParseLevel(opts.Computer); ok && opts.Computer != "" {
		m.choice.computer, m.choice.elo = true, elo
	}
	if opts.Color != "" {
		m.choice.color = opts.Color
	}
	if m.choice.color == "" {
		m.choice.color = "white"
	}
	theme := opts.Config.Theme
	if opts.Theme != "" {
		theme = opts.Theme
	}
	idx, _ := themeIndex(theme) // unknown names fall back to the default
	m.setTheme(idx)
	m.resume = m.files.loadResume()
	m.menu.restore(m.menuEntries(), opts.Config.TimeControl)
	return m
}

// Init opens the imported game, or starts a game right away when a time
// control was given.
func (m *Model) Init() tea.Cmd {
	switch {
	case len(m.importPGN) == 1:
		return m.openRecord(m.importPGN[0], m.importPath, screenMenu)
	case len(m.importPGN) > 1:
		m.showGames(m.importPGN, m.importPath, false)
	case m.startWith != nil:
		return m.startGame(*m.startWith)
	}
	return nil
}

// Update routes Bubble Tea messages to their handlers.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := m.update(msg)
	// Whatever happened may have put a position without an evaluation on
	// screen (review, the end of a game): ask the engine for one.
	return m, tea.Batch(cmd, m.autoEval())
}

// update handles one message.
func (m *Model) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tickMsg:
		return m.onTick(msg)
	case animMsg:
		return m.onAnimFrame(msg)
	case engineMsg:
		return m.onEngine(msg)
	case saveErrMsg:
		m.say(msg.what)
	case exportedMsg:
		m.onExported(msg)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return tea.Quit
		}
		m.anim = nil // any input skips the animation
		if m.help {
			return m.helpKey(msg.String())
		}
		switch m.screen {
		case screenMenu:
			return m.menuKey(msg)
		case screenThemes:
			return m.themesKey(msg.String())
		case screenOpen:
			return m.openKey(msg.String())
		case screenClocks:
			return m.clocksKey(msg.String())
		case screenOpponent:
			return m.opponentKey(msg.String())
		}
		return m.gameKey(msg.String())
	case tea.MouseMsg:
		return m.handleMouse(msg)
	}
	return nil
}

// Close stops the engine process, if one was started. Call it when the
// program ends.
func (m *Model) Close() {
	m.engine.close()
}
