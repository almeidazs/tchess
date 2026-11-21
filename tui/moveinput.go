package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/almeidazs/chess/chess"
	"github.com/almeidazs/chess/internal/sound"
)

// The move prompt (':') accepts SAN and coordinate notation and is lenient
// about how they are written: lowercase piece letters, 0-0 for castling,
// captures without the x, promotion with or without '=', extra
// disambiguation, and check or annotation marks. Exact capitalization wins;
// when only a case-insensitive reading matches and it is ambiguous between
// a pawn and a piece ("bc3": pawn or bishop), the pawn is chosen, as most
// chess apps do.

// maxInputLen bounds what the prompt accepts ("Qh4xe1=Q+" is the longest
// sensible input, with room to spare).
const maxInputLen = 12

// inputMatch is what the typed text refers to.
type inputMatch struct {
	exact  []chess.Move // moves the text fully spells
	prefix []chess.Move // moves with a spelling that starts with the text
}

// matchInput finds the legal moves the text refers to.
func matchInput(pos *chess.Position, legal []chess.Move, text string) inputMatch {
	t := normalizeInput(text)
	if t == "" {
		return inputMatch{}
	}
	if im := collectMatches(pos, legal, t, false); len(im.exact)+len(im.prefix) > 0 {
		return im
	}
	im := collectMatches(pos, legal, strings.ToLower(t), true)
	im.exact = preferPawns(pos, im.exact)
	im.prefix = preferPawns(pos, im.prefix)
	return im
}

// collectMatches compares t with every spelling of every legal move,
// lowercasing the spellings when fold is set.
func collectMatches(pos *chess.Position, legal []chess.Move, t string, fold bool) inputMatch {
	var im inputMatch
	for _, mv := range legal {
		exact, prefix := false, false
		for _, s := range spellings(pos, mv) {
			if fold {
				s = strings.ToLower(s)
			}
			exact = exact || s == t
			prefix = prefix || strings.HasPrefix(s, t)
		}
		if exact {
			im.exact = append(im.exact, mv)
		}
		if prefix {
			im.prefix = append(im.prefix, mv)
		}
	}
	return im
}

// preferPawns keeps only the pawn moves when moves mixes pawns and pieces.
func preferPawns(pos *chess.Position, moves []chess.Move) []chess.Move {
	var pawns []chess.Move
	for _, mv := range moves {
		if pos.PieceAt(mv.From()).Type() == chess.Pawn {
			pawns = append(pawns, mv)
		}
	}
	if len(pawns) == 0 {
		return moves
	}
	return pawns
}

// normalizeInput trims the text, drops check and annotation marks, and
// writes castling as O-O / O-O-O whatever the letters used.
func normalizeInput(s string) string {
	s = strings.TrimRight(strings.TrimSpace(s), "+#!?")
	switch strings.ToUpper(strings.ReplaceAll(s, "0", "O")) {
	case "O-O", "OO":
		return "O-O"
	case "O-O-O", "OOO":
		return "O-O-O"
	}
	return s
}

// spellings lists the ways mv may be written: coordinates (e2e4, e2-e4),
// and SAN with optional capture mark, optional or extra disambiguation and
// optional '=' before the promotion piece.
func spellings(pos *chess.Position, mv chess.Move) []string {
	from, to := mv.From().String(), mv.To().String()
	promo := ""
	if mv.IsPromotion() {
		promo = string(mv.Promotion().Letter())
	}
	out := []string{mv.UCI(), from + "-" + to + strings.ToLower(promo)}

	switch pc := pos.PieceAt(mv.From()); {
	case mv.Flag() == chess.FlagKingCastle:
		out = append(out, "O-O")
	case mv.Flag() == chess.FlagQueenCastle:
		out = append(out, "O-O-O")
	case pc.Type() == chess.Pawn:
		bases := []string{to}
		if mv.IsCapture() {
			bases = []string{from[:1] + "x" + to, from[:1] + to}
		}
		for _, b := range bases {
			if promo == "" {
				out = append(out, b)
			} else {
				out = append(out, b+"="+promo, b+promo)
			}
		}
	default:
		letter := string(pc.Type().Letter())
		for _, dis := range []string{"", from[:1], from[1:], from} {
			out = append(out, letter+dis+to)
			if mv.IsCapture() {
				out = append(out, letter+dis+"x"+to)
			}
		}
	}
	return out
}

// samePromotion reports whether moves are the promotion choices of a single
// pawn move (same origin and destination).
func samePromotion(moves []chess.Move) bool {
	if len(moves) == 0 {
		return false
	}
	for _, mv := range moves {
		if !mv.IsPromotion() || mv.From() != moves[0].From() || mv.To() != moves[0].To() {
			return false
		}
	}
	return true
}

// currentMatch matches the prompt's text against the current position.
func (m *Model) currentMatch() inputMatch {
	pos := m.game.Position()
	return matchInput(&pos, m.game.LegalMoves(), m.input)
}

// openInput shows the move prompt.
func (m *Model) openInput() {
	m.clearSelection()
	m.notice = ""
	m.mode = modeInput
	m.input, m.inputErr = "", ""
}

// closeInput hides the move prompt.
func (m *Model) closeInput() {
	m.mode = modePlay
	m.input, m.inputErr = "", ""
}

// inputKey handles a key while the move prompt is open.
func (m *Model) inputKey(key string) tea.Cmd {
	switch key {
	case "esc":
		m.closeInput()
	case "enter":
		return m.submitInput()
	case "backspace":
		if r := []rune(m.input); len(r) > 0 {
			m.input = string(r[:len(r)-1])
		}
		m.inputErr = ""
	default:
		if len(key) == 1 && validInputChar(key[0]) && len(m.input) < maxInputLen {
			m.input += key
			m.inputErr = ""
		}
	}
	return nil
}

// validInputChar reports whether c may appear in a typed move.
func validInputChar(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	}
	return strings.IndexByte("=-+#!?", c) >= 0
}

// submitInput plays the typed move, opens the promotion picker when only
// the piece is missing, or explains why the move was rejected.
func (m *Model) submitInput() tea.Cmd {
	if strings.TrimSpace(m.input) == "" {
		m.closeInput()
		return nil
	}
	im := m.currentMatch()
	switch {
	case len(im.exact) == 1:
		mv := im.exact[0]
		m.closeInput()
		return m.play(mv)
	case len(im.exact) == 0 && samePromotion(im.prefix):
		from := im.prefix[0].From()
		m.closeInput()
		m.selected = from
		m.targets = m.game.MovesFrom(from)
		m.promo = orderPromotions(im.prefix)
		m.promoIdx = 0
		m.mode = modePromotion
		return nil
	case len(im.exact) > 1:
		m.inputErr = "Ambíguo: " + m.sanList(im.exact)
	default:
		m.inputErr = "Lance ilegal ou incompleto"
	}
	m.sfx(sound.Error)
	return nil
}

// inputHint describes what the text currently refers to.
func (m *Model) inputHint() string {
	if m.inputErr != "" {
		return dangerStyle.Render(m.inputErr)
	}
	if strings.TrimSpace(m.input) == "" {
		return dimStyle.Render("ex.: Nf3, exd5, O-O, e8=Q, e2e4")
	}
	im := m.currentMatch()
	switch {
	case len(im.exact) == 1:
		return dimStyle.Render("→ ") + m.st.accent.Render(m.sanList(im.exact))
	case len(im.exact) > 1:
		return m.st.accent.Render("Ambíguo: " + m.sanList(im.exact))
	case samePromotion(im.prefix):
		return dimStyle.Render("promoção: enter escolhe a peça")
	case len(im.prefix) == 0:
		return dangerStyle.Render("nenhum lance legal começa assim")
	}
	return ""
}

// inputMarks returns the squares to highlight for the typed text: the
// pieces it may refer to, and the destinations once a single piece is left.
func (m *Model) inputMarks() (pieces, targets chess.Bitboard) {
	im := m.currentMatch()
	moves := im.exact
	if len(moves) == 0 {
		moves = im.prefix
	}
	for _, mv := range moves {
		pieces |= mv.From().Bitboard()
	}
	if pieces.Count() == 1 {
		for _, mv := range moves {
			targets |= mv.To().Bitboard()
		}
	}
	return pieces, targets
}

// sanList joins the moves' SAN with "ou", without check marks.
func (m *Model) sanList(moves []chess.Move) string {
	pos := m.game.Position()
	names := make([]string, 0, len(moves))
	for _, mv := range moves {
		names = append(names, strings.TrimRight(pos.SAN(mv), "+#"))
	}
	return strings.Join(names, " ou ")
}
