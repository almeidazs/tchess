package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/almeidazs/chess/internal/pixel"
)

// Theme is a board palette plus the interface's accent color. Each
// highlight has a light- and a dark-square variant so the checkered pattern
// stays visible underneath it.
type Theme struct {
	Name string

	Light, Dark               pixel.RGB // squares
	LastLight, LastDark       pixel.RGB // last move
	SelLight, SelDark         pixel.RGB // selected piece
	CursorLight, CursorDark   pixel.RGB // keyboard cursor (glyph board)
	CaptureLight, CaptureDark pixel.RGB // capturable piece (glyph board)
	Cursor                    pixel.RGB // keyboard cursor outline (sprite board)
	Mark                      pixel.RGB // legal-move dots and capture corners
	Accent                    pixel.RGB // titles, active borders, keys, picker
}

// Colors shared by every theme.
var (
	checkRGB  = pixel.Hex("#E5484D") // king in check
	pickerRGB = pixel.Hex("#E8E4DC") // promotion picker background
	hoverRGB  = pixel.Hex("#FFFFFF") // legal target under a dragged piece
	pieceFg   = lipgloss.Color("#141414")
)

// themes is the order the s key cycles through. The first is the default.
var themes = []Theme{
	{
		Name:  "Madeira",
		Light: pixel.Hex("#F0D9B5"), Dark: pixel.Hex("#B58863"),
		LastLight: pixel.Hex("#CDD26A"), LastDark: pixel.Hex("#AAA23A"),
		SelLight: pixel.Hex("#9DB86A"), SelDark: pixel.Hex("#7A9A47"),
		CursorLight: pixel.Hex("#A9C8EC"), CursorDark: pixel.Hex("#7AA2D2"),
		CaptureLight: pixel.Hex("#E8A792"), CaptureDark: pixel.Hex("#C8775E"),
		Cursor: pixel.Hex("#3D7DD8"),
		Mark:   pixel.Hex("#14551E"),
		Accent: pixel.Hex("#E2B86B"),
	},
	derivedTheme("Floresta", palette{
		light: "#EEEED2", dark: "#769656",
		last: "#F6F669", sel: "#C7D84A", cursor: "#3D7DD8",
		mark: "#1F3A12", accent: "#A3C76D",
	}),
	derivedTheme("Oceano", palette{
		light: "#DEE3E6", dark: "#8CA2AD",
		last: "#E8DC6A", sel: "#7FC97F", cursor: "#F28C28",
		mark: "#123047", accent: "#7FB2D9",
	}),
	derivedTheme("Grafite", palette{
		light: "#C9C9C9", dark: "#8C8C8C",
		last: "#E8C65A", sel: "#8FC46A", cursor: "#3D7DD8",
		mark: "#262626", accent: "#C8CDD6",
	}),
	derivedTheme("Coral", palette{
		light: "#F4E1D2", dark: "#D9846F",
		last: "#F2D45C", sel: "#9CC46A", cursor: "#3D7DD8",
		mark: "#5A1E12", accent: "#F2937C",
	}),
	derivedTheme("Mármore", palette{
		light: "#F7F7F4", dark: "#A3A8AE",
		last: "#F0D04A", sel: "#7FC46A", cursor: "#2F6FD0",
		mark: "#1B1D21", accent: "#E9EBEE",
	}),
}

// palette is the handful of colors a theme is derived from.
type palette struct {
	light, dark, last, sel, cursor, mark, accent string
}

// derivedTheme builds a theme from a palette, tinting both square colors
// with each highlight color.
func derivedTheme(name string, p palette) Theme {
	l, d := pixel.Hex(p.light), pixel.Hex(p.dark)
	tint := func(hex string, t float64) (pixel.RGB, pixel.RGB) {
		c := pixel.Hex(hex)
		return pixel.Blend(l, c, t), pixel.Blend(d, c, t)
	}
	th := Theme{Name: name, Light: l, Dark: d, Cursor: pixel.Hex(p.cursor), Mark: pixel.Hex(p.mark), Accent: pixel.Hex(p.accent)}
	th.LastLight, th.LastDark = tint(p.last, 0.55)
	th.SelLight, th.SelDark = tint(p.sel, 0.6)
	th.CursorLight, th.CursorDark = tint(p.cursor, 0.45)
	th.CaptureLight, th.CaptureDark = tint("#E0533D", 0.45)
	return th
}

// themeIndex finds a theme by name, ignoring case and accents ("marmore"
// matches "Mármore").
func themeIndex(name string) (int, bool) {
	want := foldName(name)
	for i, t := range themes {
		if foldName(t.Name) == want {
			return i, true
		}
	}
	return 0, false
}

// HasTheme reports whether name is a theme, ignoring case and accents.
func HasTheme(name string) bool {
	_, ok := themeIndex(name)
	return ok
}

// ThemeNames lists the available themes, in cycling order.
func ThemeNames() []string {
	out := make([]string, len(themes))
	for i, t := range themes {
		out[i] = t.Name
	}
	return out
}

// accentFolder strips the Portuguese diacritics.
var accentFolder = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "é", "e", "ê", "e", "í", "i",
	"ó", "o", "ô", "o", "õ", "o", "ú", "u", "ç", "c",
)

// foldName lowercases s and strips its diacritics.
func foldName(s string) string {
	return accentFolder.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// shade picks the light- or dark-square variant of a color.
func shade(light bool, l, d pixel.RGB) pixel.RGB {
	if light {
		return l
	}
	return d
}

// Neutral interface palette, the same in every theme.
var (
	subtle  = lipgloss.Color("#6C6F7E")
	text    = lipgloss.Color("#E6E6E6")
	danger  = lipgloss.Color("#FF6B6B")
	success = lipgloss.Color("#7BD88F")
	panelBg = lipgloss.Color("#2A2B36")

	dimStyle    = lipgloss.NewStyle().Foreground(subtle)
	textStyle   = lipgloss.NewStyle().Foreground(text)
	dangerStyle = lipgloss.NewStyle().Bold(true).Foreground(danger)

	capturedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D8D2C6"))
	whiteBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EDE7DA"))
	blackBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C7A86"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle).
			Padding(0, 1)
)

// styles are the interface styles that follow the theme's accent color.
type styles struct {
	accentColor lipgloss.Color
	title       lipgloss.Style
	accent      lipgloss.Style
	key         lipgloss.Style
	activeBox   lipgloss.Style
	menuBox     lipgloss.Style
}

func newStyles(t *Theme) styles {
	c := lipgloss.Color(t.Accent.Hex())
	bold := lipgloss.NewStyle().Bold(true).Foreground(c)
	return styles{
		accentColor: c,
		title:       bold,
		accent:      bold,
		key:         bold,
		activeBox:   boxStyle.BorderForeground(c),
		menuBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c).
			Padding(menuPadY, menuPadX),
	}
}

// theme is the active theme.
func (m *Model) theme() *Theme { return &themes[m.themeIdx] }

// setTheme activates the theme at index i.
func (m *Model) setTheme(i int) {
	m.themeIdx = i
	m.st = newStyles(&themes[i])
}
