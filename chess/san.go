package chess

import (
	"fmt"
	"strings"
)

// SAN returns the Standard Algebraic Notation of m, which must be legal in p
// (e.g. "Nbd7", "exd6", "e8=Q+", "O-O-O#"). p is temporarily modified to
// determine the check suffix, and restored before returning.
func (p *Position) SAN(m Move) string {
	var legal MoveList
	p.LegalMoves(&legal)
	return p.san(m, &legal)
}

// san is SAN with the legal moves already generated.
func (p *Position) san(m Move, legal *MoveList) string {
	var sb strings.Builder
	from, to := m.From(), m.To()
	pc := p.board[from]

	switch {
	case m.Flag() == FlagKingCastle:
		sb.WriteString("O-O")
	case m.Flag() == FlagQueenCastle:
		sb.WriteString("O-O-O")
	case pc.Type() == Pawn:
		if m.IsCapture() {
			sb.WriteByte(byte('a' + from.File()))
			sb.WriteByte('x')
		}
		sb.WriteString(to.String())
		if m.IsPromotion() {
			sb.WriteByte('=')
			sb.WriteByte(m.Promotion().Letter())
		}
	default:
		sb.WriteByte(pc.Type().Letter())
		ambiguous, sameFile, sameRank := false, false, false
		for _, o := range legal.Slice() {
			if o == m || o.To() != to || p.board[o.From()] != pc {
				continue
			}
			ambiguous = true
			sameFile = sameFile || o.From().File() == from.File()
			sameRank = sameRank || o.From().Rank() == from.Rank()
		}
		if ambiguous {
			switch {
			case !sameFile:
				sb.WriteByte(byte('a' + from.File()))
			case !sameRank:
				sb.WriteByte(byte('1' + from.Rank()))
			default:
				sb.WriteString(from.String())
			}
		}
		if m.IsCapture() {
			sb.WriteByte('x')
		}
		sb.WriteString(to.String())
	}

	u := p.Make(m)
	if p.InCheck() {
		var reply MoveList
		p.LegalMoves(&reply)
		if reply.n == 0 {
			sb.WriteByte('#')
		} else {
			sb.WriteByte('+')
		}
	}
	p.Unmake(m, u)
	return sb.String()
}

// ParseSAN finds the legal move written in SAN. Check/mate marks and
// annotations (+, #, !, ?) are optional; "0-0" is accepted for castling.
func (p *Position) ParseSAN(s string) (Move, error) {
	want := normalizeSAN(s)
	var legal MoveList
	p.LegalMoves(&legal)
	for _, m := range legal.Slice() {
		if normalizeSAN(p.san(m, &legal)) == want {
			return m, nil
		}
	}
	return NullMove, fmt.Errorf("chess: illegal or ambiguous move %q", s)
}

// normalizeSAN strips check marks and annotations and accepts "0-0".
func normalizeSAN(s string) string {
	s = strings.TrimRight(strings.TrimSpace(s), "+#!?")
	return strings.ReplaceAll(s, "0", "O")
}

// ParseUCI finds the legal move written in coordinate notation ("e2e4", "e7e8q").
func (p *Position) ParseUCI(s string) (Move, error) {
	var legal MoveList
	p.LegalMoves(&legal)
	for _, m := range legal.Slice() {
		if m.UCI() == s {
			return m, nil
		}
	}
	return NullMove, fmt.Errorf("chess: illegal move %q", s)
}
