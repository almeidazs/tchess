package tui

import (
	"github.com/almeidazs/chess/chess"
)

// User-facing strings (Portuguese). The engine only exposes enums; all
// wording lives here.

// colorName is the side's name in Portuguese.
func colorName(c chess.Color) string {
	if c == chess.White {
		return "Brancas"
	}
	return "Pretas"
}

// terminationText explains how a game ended.
func terminationText(t chess.Termination) string {
	switch t {
	case chess.Checkmate:
		return "xeque-mate"
	case chess.Stalemate:
		return "afogamento"
	case chess.InsufficientMaterial:
		return "material insuficiente"
	case chess.FivefoldRepetition:
		return "repetição quíntupla"
	case chess.SeventyFiveMoveRule:
		return "regra dos 75 lances"
	case chess.ThreefoldRepetition:
		return "repetição tripla"
	case chess.FiftyMoveRule:
		return "regra dos 50 lances"
	case chess.DrawAgreement:
		return "acordo mútuo"
	case chess.Resignation:
		return "desistência"
	case chess.Timeout:
		return "tempo esgotado"
	case chess.TimeoutVsInsufficientMaterial:
		return "tempo esgotado, sem material para mate"
	}
	return ""
}

// outcomeHeadline announces the winner or the draw.
func outcomeHeadline(o chess.Outcome) string {
	switch o {
	case chess.WhiteWins:
		return "Brancas vencem"
	case chess.BlackWins:
		return "Pretas vencem"
	case chess.Draw:
		return "Empate"
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
