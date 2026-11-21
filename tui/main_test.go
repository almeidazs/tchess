package tui

import (
	"os"
	"testing"

	"github.com/almeidazs/tchess/uci/ucitest"
)

// TestMain lets tests start this binary as a fake UCI engine.
func TestMain(m *testing.M) {
	ucitest.Main()
	os.Exit(m.Run())
}
