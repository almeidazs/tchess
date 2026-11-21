package chess

import (
	"testing"
)

func BenchmarkPerftInitial4(b *testing.B) {
	p := StartingPosition()
	b.ResetTimer()
	for b.Loop() {
		p.Perft(4)
	}
}

func BenchmarkPerftKiwipete3(b *testing.B) {
	p, _ := ParseFEN(perftSuite[1].fen)
	b.ResetTimer()
	for b.Loop() {
		p.Perft(3)
	}
}

func BenchmarkLegalMovesKiwipete(b *testing.B) {
	p, _ := ParseFEN(perftSuite[1].fen)
	var ml MoveList
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		p.LegalMoves(&ml)
	}
}

func BenchmarkMakeUnmake(b *testing.B) {
	p, _ := ParseFEN(perftSuite[1].fen)
	var ml MoveList
	p.LegalMoves(&ml)
	moves := ml.Slice()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		m := moves[i%len(moves)]
		u := p.Make(m)
		p.Unmake(m, u)
	}
}
