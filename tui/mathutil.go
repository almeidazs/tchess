package tui

// abs returns |x|.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// sq returns x².
func sq(x int) int { return x * x }
