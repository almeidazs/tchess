package tui

import (
	"github.com/almeidazs/tchess/chess"
)

// User-facing strings. The engine only exposes enums; all
// wording lives here.

// colorName is the side's name.
func colorName(c chess.Color) string {
	if c == chess.White {
		return "White"
	}
	return "Black"
}

// terminationText explains how a game ended.
func terminationText(t chess.Termination) string {
	switch t {
	case chess.Checkmate:
		return "checkmate"
	case chess.Stalemate:
		return "stalemate"
	case chess.InsufficientMaterial:
		return "insufficient material"
	case chess.FivefoldRepetition:
		return "fivefold repetition"
	case chess.SeventyFiveMoveRule:
		return "75-move rule"
	case chess.ThreefoldRepetition:
		return "threefold repetition"
	case chess.FiftyMoveRule:
		return "50-move rule"
	case chess.DrawAgreement:
		return "agreement"
	case chess.Resignation:
		return "resignation"
	case chess.Timeout:
		return "time forfeit"
	case chess.TimeoutVsInsufficientMaterial:
		return "time forfeit, no mating material"
	case chess.RecordedResult:
		return "result from the PGN"
	}
	return ""
}

// legacyTerminations are the reasons older versions wrote, in Portuguese,
// in the Termination tag of the games they saved.
var legacyTerminations = map[string]chess.Termination{
	"xeque-mate":                             chess.Checkmate,
	"afogamento":                             chess.Stalemate,
	"material insuficiente":                  chess.InsufficientMaterial,
	"repetição quíntupla":                    chess.FivefoldRepetition,
	"regra dos 75 lances":                    chess.SeventyFiveMoveRule,
	"repetição tripla":                       chess.ThreefoldRepetition,
	"regra dos 50 lances":                    chess.FiftyMoveRule,
	"acordo mútuo":                           chess.DrawAgreement,
	"desistência":                            chess.Resignation,
	"tempo esgotado":                         chess.Timeout,
	"tempo esgotado, sem material para mate": chess.TimeoutVsInsufficientMaterial,
	"resultado do PGN":                       chess.RecordedResult,
}

// terminationFromText reads back a reason written by terminationText (or
// by an older version), as in the Termination tag of the games tchess
// saves. Anything else is RecordedResult.
func terminationFromText(s string) chess.Termination {
	for t := chess.Checkmate; t <= chess.RecordedResult; t++ {
		if s != "" && terminationText(t) == s {
			return t
		}
	}
	if t, ok := legacyTerminations[s]; ok {
		return t
	}
	return chess.RecordedResult
}

// outcomeHeadline announces the winner or the draw.
func outcomeHeadline(o chess.Outcome) string {
	switch o {
	case chess.WhiteWins:
		return "White wins"
	case chess.BlackWins:
		return "Black wins"
	case chess.Draw:
		return "Draw"
	}
	return ""
}

// outcomeScore is the score in chess notation.
func outcomeScore(o chess.Outcome) string {
	switch o {
	case chess.WhiteWins:
		return "1-0"
	case chess.BlackWins:
		return "0-1"
	case chess.Draw:
		return "½-½"
	}
	return "*"
}
