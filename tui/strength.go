package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/almeidazs/tchess/uci"
)

// level is a named computer strength. elo 0 means full strength.
type level struct {
	name string
	elo  int
}

// levels are the presets offered in the menu, weakest first.
var levels = []level{
	{"Beginner", 800},
	{"Easy", 1200},
	{"Medium", 1600},
	{"Hard", 2000},
	{"Expert", 2400},
	{"Max", 0},
}

// Custom Elo bounds accepted from the player.
const (
	minElo = 400
	maxElo = 3500
)

// levelName names a strength: a preset's name, or "Elo 1750".
func levelName(elo int) string {
	for _, l := range levels {
		if l.elo == elo {
			return l.name
		}
	}
	return "Elo " + strconv.Itoa(elo)
}

// ParseLevel reads a computer strength: a preset name ("medium", "max")
// or an Elo number. elo 0 means full strength.
func ParseLevel(s string) (elo int, ok bool) {
	s = strings.TrimSpace(s)
	for _, l := range levels {
		if strings.EqualFold(l.name, s) {
			return l.elo, true
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < minElo || n > maxElo {
		return 0, false
	}
	return n, true
}

// strength is how an engine is told to play at a given Elo.
type strength struct {
	options [][2]string   // setoption name/value pairs, in order
	depth   int           // search depth limit, 0 for none
	think   time.Duration // thinking time: the cap with a clock, the move time without
}

// key identifies the option set, so options are only resent on change.
func (s strength) key() string { return fmt.Sprint(s.options) }

// skillTable is how weak levels are played with "Skill Level" alone,
// below what UCI_Elo allows (Stockfish's minimum is 1320): skill, depth
// and a short thinking time, as Lichess does for its weakest AI levels.
var skillTable = []struct {
	elo, skill, depth int
	think             time.Duration
}{
	{800, 3, 1, 50 * time.Millisecond},
	{1100, 6, 2, 100 * time.Millisecond},
	{1400, 9, 3, 150 * time.Millisecond},
	{1700, 11, 4, 200 * time.Millisecond},
	{2000, 14, 6, 300 * time.Millisecond},
	{2300, 17, 8, 400 * time.Millisecond},
	{2700, 20, 10, 500 * time.Millisecond},
}

// thinkFor is the thinking time of a strength: short for weak levels so
// the computer answers quickly, longer as it gets stronger.
func thinkFor(elo int) time.Duration {
	switch {
	case elo == 0:
		return 0 // full strength: managed by the clock, or 3 s without one
	case elo <= 1000:
		return 300 * time.Millisecond
	case elo <= 1400:
		return 600 * time.Millisecond
	case elo <= 1800:
		return time.Second
	case elo <= 2200:
		return 2 * time.Second
	}
	return 3 * time.Second
}

// fullStrengthMoveTime is how long the strongest level thinks per move in
// a game without a clock.
const fullStrengthMoveTime = 3 * time.Second

// strengthFor works out how to make engine e play at elo, using the
// options it declares:
//
//   - UCI_LimitStrength and UCI_Elo when elo is within the engine's range;
//   - "Skill Level" with a depth limit below that range (or without it);
//   - only a depth limit for engines with neither.
func strengthFor(e map[string]uci.Option, elo int) strength {
	s := strength{think: thinkFor(elo)}
	_, hasLimit := e["uci_limitstrength"]
	eloOpt, hasElo := e["uci_elo"]
	skill, hasSkill := e["skill level"]

	if elo == 0 { // full strength
		if hasLimit {
			s.options = append(s.options, [2]string{"UCI_LimitStrength", "false"})
		}
		if hasSkill {
			s.options = append(s.options, [2]string{"Skill Level", strconv.Itoa(skill.Max)})
		}
		return s
	}
	if hasLimit && hasElo && elo >= eloOpt.Min {
		s.options = append(s.options,
			[2]string{"UCI_LimitStrength", "true"},
			[2]string{"UCI_Elo", strconv.Itoa(min(elo, eloOpt.Max))})
		if hasSkill {
			s.options = append(s.options, [2]string{"Skill Level", strconv.Itoa(skill.Max)})
		}
		return s
	}
	row := skillTable[len(skillTable)-1]
	for _, r := range skillTable {
		if elo <= r.elo {
			row = r
			break
		}
	}
	if hasLimit {
		s.options = append(s.options, [2]string{"UCI_LimitStrength", "false"})
	}
	if hasSkill {
		v := min(max(row.skill, skill.Min), skill.Max)
		s.options = append(s.options, [2]string{"Skill Level", strconv.Itoa(v)})
	}
	s.depth = row.depth
	s.think = min(s.think, row.think*4) // weak levels answer fast
	return s
}
