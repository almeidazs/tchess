package chess

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// StartFEN is the standard initial position.
const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// ParseFEN parses a position in Forsyth-Edwards Notation. The halfmove and
// fullmove fields are optional. The position is validated (one king per
// side, no pawns on the back ranks, side not to move not in check) and
// normalized: castling rights without the king and rook in place are dropped
// and an en passant square no pawn can capture on is ignored.
func ParseFEN(fen string) (*Position, error) {
	fields := strings.Fields(fen)
	if len(fields) != 4 && len(fields) != 6 {
		return nil, fmt.Errorf("chess: FEN needs 4 or 6 fields, got %d", len(fields))
	}
	p := &Position{ep: NoSquare, fullmove: 1}
	for i := range p.board {
		p.board[i] = NoPiece
	}

	ranks := strings.Split(fields[0], "/")
	if len(ranks) != 8 {
		return nil, errors.New("chess: FEN board must have 8 ranks")
	}
	for i, row := range ranks {
		r, f := 7-i, 0
		for _, ch := range row {
			switch {
			case ch >= '1' && ch <= '8':
				f += int(ch - '0')
			default:
				idx := strings.IndexRune(fenPieces, ch)
				if idx < 0 {
					return nil, fmt.Errorf("chess: invalid piece %q in FEN", ch)
				}
				if f > 7 {
					return nil, fmt.Errorf("chess: rank %d overflows in FEN", r+1)
				}
				p.put(Piece(idx), NewSquare(f, r))
				f++
			}
		}
		if f != 8 {
			return nil, fmt.Errorf("chess: rank %d has %d files in FEN", r+1, f)
		}
	}

	switch fields[1] {
	case "w":
		p.side = White
	case "b":
		p.side = Black
	default:
		return nil, fmt.Errorf("chess: invalid side to move %q", fields[1])
	}

	if fields[2] != "-" {
		for _, ch := range fields[2] {
			switch ch {
			case 'K':
				p.castling |= WhiteKingside
			case 'Q':
				p.castling |= WhiteQueenside
			case 'k':
				p.castling |= BlackKingside
			case 'q':
				p.castling |= BlackQueenside
			default:
				return nil, fmt.Errorf("chess: invalid castling field %q", fields[2])
			}
		}
	}

	if fields[3] != "-" {
		sq, err := ParseSquare(fields[3])
		if err != nil {
			return nil, err
		}
		p.ep = sq
	}

	if len(fields) == 6 {
		hm, err1 := strconv.ParseUint(fields[4], 10, 16)
		fm, err2 := strconv.ParseUint(fields[5], 10, 16)
		if err1 != nil || err2 != nil || fm == 0 {
			return nil, errors.New("chess: invalid move counters in FEN")
		}
		p.halfmove, p.fullmove = uint16(hm), uint16(fm)
	}

	if err := p.validate(); err != nil {
		return nil, err
	}
	p.normalize()
	p.hash = p.computeHash()
	return p, nil
}

// validate rejects positions that cannot arise in a legal game.
func (p *Position) validate() error {
	for c := White; c <= Black; c++ {
		if p.pieces[c][King].Count() != 1 {
			return fmt.Errorf("chess: %s must have exactly one king", c)
		}
	}
	if (p.pieces[White][Pawn]|p.pieces[Black][Pawn])&(rank1|rank8) != 0 {
		return errors.New("chess: pawns cannot stand on the first or last rank")
	}
	them := p.side.Other()
	if p.attackersTo(p.KingSquare(them), p.all)&p.occ[p.side] != 0 {
		return errors.New("chess: the side not to move is in check")
	}
	return nil
}

// normalize drops castling rights and an en passant square that the
// position cannot actually use.
func (p *Position) normalize() {
	need := []struct {
		right      CastlingRights
		king, rook Piece
		ks, rs     Square
	}{
		{WhiteKingside, MakePiece(White, King), MakePiece(White, Rook), E1, H1},
		{WhiteQueenside, MakePiece(White, King), MakePiece(White, Rook), E1, A1},
		{BlackKingside, MakePiece(Black, King), MakePiece(Black, Rook), E8, H8},
		{BlackQueenside, MakePiece(Black, King), MakePiece(Black, Rook), E8, A8},
	}
	for _, n := range need {
		if p.board[n.ks] != n.king || p.board[n.rs] != n.rook {
			p.castling &^= n.right
		}
	}

	if p.ep != NoSquare {
		us, them := p.side, p.side.Other()
		wantRank, pushed := 5, p.ep-8
		if us == Black {
			wantRank, pushed = 2, p.ep+8
		}
		valid := p.ep.Rank() == wantRank &&
			p.board[p.ep] == NoPiece &&
			p.board[pushed] == MakePiece(them, Pawn) &&
			pawnAttacks[them][p.ep]&p.pieces[us][Pawn] != 0
		if !valid {
			p.ep = NoSquare
		}
	}
}

// FEN returns the position in Forsyth-Edwards Notation.
func (p *Position) FEN() string {
	var sb strings.Builder
	for r := 7; r >= 0; r-- {
		empty := 0
		for f := range 8 {
			pc := p.board[NewSquare(f, r)]
			if pc == NoPiece {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteByte(byte('0' + empty))
				empty = 0
			}
			sb.WriteByte(fenPieces[pc])
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		if r > 0 {
			sb.WriteByte('/')
		}
	}
	if p.side == White {
		sb.WriteString(" w ")
	} else {
		sb.WriteString(" b ")
	}
	if p.castling == NoCastling {
		sb.WriteByte('-')
	} else {
		for i, ch := range "KQkq" {
			if p.castling&(1<<i) != 0 {
				sb.WriteRune(ch)
			}
		}
	}
	fmt.Fprintf(&sb, " %s %d %d", p.ep, p.halfmove, p.fullmove)
	return sb.String()
}
