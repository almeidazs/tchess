package chess

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Tag is a PGN tag pair, such as [Event "Casual game"].
type Tag struct {
	Name, Value string
}

// pgnLineWidth is the maximum movetext line length when writing PGN.
const pgnLineWidth = 79

// PGN writes the game in Portable Game Notation.
//
// tags are written in the order given, except that a Result tag always
// carries the game's actual result (and is added when missing), and SetUp
// and FEN tags are added when the game did not start from the initial
// position. comments, when not nil, holds the comment written after each
// move, indexed by ply ("" for none). A comment may not contain '}'; any is
// replaced by ')'.
func (g *Game) PGN(tags []Tag, comments []string) string {
	var sb strings.Builder
	result := g.outcome.String()
	hasResult := false
	for _, t := range tags {
		switch t.Name {
		case "SetUp", "FEN":
			continue
		case "Result":
			t.Value, hasResult = result, true
		}
		writeTag(&sb, t)
	}
	if !hasResult {
		writeTag(&sb, Tag{"Result", result})
	}
	if fen := g.start.FEN(); fen != StartFEN {
		writeTag(&sb, Tag{"SetUp", "1"})
		writeTag(&sb, Tag{"FEN", fen})
	}
	sb.WriteByte('\n')

	var tokens []string
	number, side := g.start.FullmoveNumber(), g.start.SideToMove()
	needNumber := true // a black move needs "N..." at the start and after a comment
	for i, r := range g.history {
		switch {
		case side == White:
			tokens = append(tokens, strconv.Itoa(number)+".")
		case needNumber:
			tokens = append(tokens, strconv.Itoa(number)+"...")
		}
		tokens = append(tokens, r.SAN)
		needNumber = false
		if i < len(comments) && comments[i] != "" {
			tokens = append(tokens, "{"+strings.ReplaceAll(comments[i], "}", ")")+"}")
			needNumber = true
		}
		if side == Black {
			number++
		}
		side = side.Other()
	}
	tokens = append(tokens, result)

	line := 0
	for i, tok := range tokens {
		if i > 0 {
			if line+1+len(tok) > pgnLineWidth {
				sb.WriteByte('\n')
				line = 0
			} else {
				sb.WriteByte(' ')
				line++
			}
		}
		sb.WriteString(tok)
		line += len(tok)
	}
	sb.WriteByte('\n')
	return sb.String()
}

// writeTag writes one tag pair, escaping quotes and backslashes.
func writeTag(sb *strings.Builder, t Tag) {
	v := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(t.Value)
	fmt.Fprintf(sb, "[%s \"%s\"]\n", t.Name, v)
}

// PGNGame is a game read from PGN.
type PGNGame struct {
	// Game is the game replayed from the record. It ends on its own when
	// the moves do (checkmate, stalemate...); a result the moves do not
	// explain, such as a resignation, is only in Result, for the caller to
	// apply with Game.Record and the reason it knows.
	Game *Game
	// Tags are the tag pairs, in file order.
	Tags []Tag
	// Comments holds the comment after each move, indexed by ply ("" for
	// none). Several comments after one move are joined with a space.
	Comments []string
	// Result is the result written in the record (Ongoing for "*").
	Result Outcome
}

// Tag returns the value of the named tag, or "".
func (pg *PGNGame) Tag(name string) string {
	for _, t := range pg.Tags {
		if t.Name == name {
			return t.Value
		}
	}
	return ""
}

// ParsePGN reads the first game of a PGN text. It accepts what other
// programs write: comments ({...} and ;), variations (skipped), NAGs,
// move numbers with or without a space, and SAN or coordinate moves.
func ParsePGN(text string) (*PGNGame, error) {
	chunks := splitPGN(text)
	if len(chunks) == 0 {
		return parseGame("")
	}
	return parseGame(chunks[0])
}

// ParsePGNGames reads every game of a PGN text, in order. A game that
// cannot be read is left out; the error then tells which ones (numbered
// from 1) and why, while the games returned are still usable.
func ParsePGNGames(text string) ([]*PGNGame, error) {
	var games []*PGNGame
	var errs []error
	for i, chunk := range splitPGN(text) {
		pg, err := parseGame(chunk)
		if err != nil {
			errs = append(errs, fmt.Errorf("game %d: %w", i+1, err))
			continue
		}
		games = append(games, pg)
	}
	return games, errors.Join(errs...)
}

// splitPGN cuts a PGN text into games: a new game starts at a tag line
// that follows movetext. Lines inside a {comment} never count as tags.
func splitPGN(text string) []string {
	text = strings.TrimPrefix(text, "\ufeff") // byte order mark
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var chunks []string
	var cur strings.Builder
	inMovetext, inComment := false, false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inComment && strings.HasPrefix(trimmed, "[") && inMovetext {
			chunks = append(chunks, cur.String())
			cur.Reset()
			inMovetext = false
		}
		if trimmed != "" && !inComment && !strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(line, "%") {
			inMovetext = true
		}
	scan:
		for _, c := range line {
			switch {
			case c == '{':
				inComment = true
			case c == '}':
				inComment = false
			case c == ';' && !inComment:
				break scan // the rest of the line is a comment
			}
		}
		cur.WriteString(line)
		cur.WriteByte('\n')
	}
	if strings.TrimSpace(cur.String()) != "" {
		chunks = append(chunks, cur.String())
	}
	return chunks
}

// parseGame reads one game: its tags, then its movetext.
func parseGame(text string) (*PGNGame, error) {
	pg := &PGNGame{}
	lines := strings.Split(text, "\n")

	// Tag section: every '[' line before the movetext.
	i := 0
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			if len(pg.Tags) > 0 {
				i++
				break
			}
			continue
		}
		if !strings.HasPrefix(line, "[") {
			break
		}
		t, err := parseTag(line)
		if err != nil {
			return nil, err
		}
		pg.Tags = append(pg.Tags, t)
	}

	fen := StartFEN
	if f := pg.Tag("FEN"); f != "" {
		fen = f
	}
	g, err := NewGameFromFEN(fen)
	if err != nil {
		return nil, fmt.Errorf("chess: PGN FEN tag: %w", err)
	}
	pg.Game = g
	pg.Result = parseResult(pg.Tag("Result"))

	var mt strings.Builder
	for ; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "%") { // escape line
			continue
		}
		mt.WriteString(lines[i])
		mt.WriteByte('\n')
	}
	if err := pg.readMovetext(mt.String()); err != nil {
		return nil, err
	}

	for len(pg.Comments) < g.Ply() {
		pg.Comments = append(pg.Comments, "")
	}
	return pg, nil
}

// readMovetext plays the moves and collects the comments.
func (pg *PGNGame) readMovetext(s string) error {
	addComment := func(c string) {
		c = strings.Join(strings.Fields(c), " ")
		ply := pg.Game.Ply() - 1
		if ply < 0 || c == "" {
			return // a comment before the first move describes the game
		}
		for len(pg.Comments) <= ply {
			pg.Comments = append(pg.Comments, "")
		}
		if pg.Comments[ply] != "" {
			pg.Comments[ply] += " "
		}
		pg.Comments[ply] += c
	}

	for i := 0; i < len(s); {
		switch s[i] {
		case ' ', '\t', '\n', '\r', '.':
			i++
		case '{':
			end := strings.IndexByte(s[i:], '}')
			if end < 0 {
				return fmt.Errorf("chess: PGN comment not closed")
			}
			addComment(s[i+1 : i+end])
			i += end + 1
		case ';':
			end := strings.IndexByte(s[i:], '\n')
			if end < 0 {
				end = len(s) - i
			}
			addComment(s[i+1 : i+end])
			i += end
		case '(':
			n, err := skipVariation(s[i:])
			if err != nil {
				return err
			}
			i += n
		case '$': // numeric annotation glyph
			i++
			for i < len(s) && s[i] >= '0' && s[i] <= '9' {
				i++
			}
		default:
			j := i
			for j < len(s) && !strings.ContainsRune(" \t\n\r{};($", rune(s[j])) {
				j++
			}
			tok := s[i:j]
			i = j
			if done, err := pg.readToken(tok); err != nil || done {
				return err
			}
		}
	}
	return nil
}

// readToken handles a move number, a move or a result. It reports whether
// the token ended the game.
func (pg *PGNGame) readToken(tok string) (bool, error) {
	if r := parseResult(tok); r != Ongoing || tok == "*" {
		pg.Result = r
		return true, nil
	}
	// A move number, possibly glued to the move: "12.", "12...", "12.Nf3".
	k := 0
	for k < len(tok) && tok[k] >= '0' && tok[k] <= '9' {
		k++
	}
	if k > 0 && k < len(tok) && tok[k] == '.' {
		tok = strings.TrimLeft(tok[k:], ".")
	} else if k == len(tok) {
		return false, nil // a bare number
	}
	tok = strings.TrimRight(tok, "!?")
	if tok == "" || tok == "e.p." {
		return false, nil
	}
	if pg.Game.IsOver() {
		return false, fmt.Errorf("chess: PGN move %q after the game ended", tok)
	}
	pos := pg.Game.Position()
	m, err := pos.ParseSAN(tok)
	if err != nil {
		if m2, err2 := pos.ParseUCI(strings.ToLower(tok)); err2 == nil {
			m, err = m2, nil
		}
	}
	if err != nil {
		return false, fmt.Errorf("chess: PGN move %d (%q) is not legal", pg.Game.Ply()+1, tok)
	}
	return false, pg.Game.Play(m)
}

// skipVariation returns the length of the parenthesized variation at the
// start of s, including nested variations and comments.
func skipVariation(s string) (int, error) {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i + 1, nil
			}
		case '{':
			end := strings.IndexByte(s[i:], '}')
			if end < 0 {
				return 0, fmt.Errorf("chess: PGN comment not closed")
			}
			i += end
		}
	}
	return 0, fmt.Errorf("chess: PGN variation not closed")
}

// parseTag reads `[Name "Value"]`.
func parseTag(line string) (Tag, error) {
	inner := strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
	name, rest, ok := strings.Cut(strings.TrimSpace(inner), " ")
	rest = strings.TrimSpace(rest)
	if !ok || len(rest) < 2 || rest[0] != '"' || rest[len(rest)-1] != '"' {
		return Tag{}, fmt.Errorf("chess: PGN tag %q is malformed", line)
	}
	v := strings.NewReplacer(`\\`, `\`, `\"`, `"`).Replace(rest[1 : len(rest)-1])
	return Tag{name, v}, nil
}

// parseResult reads a game termination marker.
func parseResult(s string) Outcome {
	switch s {
	case "1-0":
		return WhiteWins
	case "0-1":
		return BlackWins
	case "1/2-1/2", "½-½":
		return Draw
	}
	return Ongoing
}
