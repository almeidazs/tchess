package sound

import (
	"bytes"
	"encoding/binary"
	"math"
)

// SampleRate is the rate of the generated audio, in Hz (mono, 16-bit).
const SampleRate = 22050

// peak is the level every effect is normalized to, as a fraction of full
// scale, so they all sound equally loud.
const peak = 0.7

// buffer is a mono signal in [-1, 1] being composed.
type buffer []float64

func newBuffer(seconds float64) buffer {
	return make(buffer, int(seconds*SampleRate))
}

// tone adds a sine at freq Hz starting at the given offset (seconds), with
// a short attack and an exponential decay of time constant tau.
func (b buffer) tone(at, freq, tau, amp float64) {
	const attack = 0.002
	start := int(at * SampleRate)
	for i := start; i < len(b); i++ {
		t := float64(i-start) / SampleRate
		env := math.Exp(-t / tau)
		if t < attack {
			env *= t / attack
		}
		if env < 1e-4 && t > attack {
			break
		}
		b[i] += amp * env * math.Sin(2*math.Pi*freq*t)
	}
}

// bell adds a tone with a quieter octave on top, which reads as a chime.
func (b buffer) bell(at, freq, tau, amp float64) {
	b.tone(at, freq, tau, amp)
	b.tone(at, 2*freq, tau/2, amp*0.3)
	b.tone(at, 3*freq, tau/3, amp*0.1)
}

// click adds a burst of low-passed noise: the attack of a wooden knock.
// Noise comes from a fixed-seed generator, so the output is reproducible.
func (b buffer) click(at, tau, amp float64) {
	start := int(at * SampleRate)
	seed := uint32(0x9E3779B9)
	var lp float64
	for i := start; i < len(b); i++ {
		t := float64(i-start) / SampleRate
		env := math.Exp(-t / tau)
		if env < 1e-4 {
			break
		}
		seed ^= seed << 13
		seed ^= seed >> 17
		seed ^= seed << 5
		white := float64(seed)/math.MaxUint32*2 - 1
		lp += 0.35 * (white - lp)
		b[i] += amp * env * lp
	}
}

// knock adds the sound of a piece set down on a wooden board: a short
// resonant body plus the noise of the impact. Lower pitch sounds heavier.
func (b buffer) knock(at, pitch, amp float64) {
	b.tone(at, pitch, 0.018, amp)
	b.tone(at, pitch*2.76, 0.007, amp*0.45)
	b.click(at, 0.003, amp*0.6)
}

// normalize scales the signal so its loudest sample hits peak.
func (b buffer) normalize() {
	var m float64
	for _, v := range b {
		m = max(m, math.Abs(v))
	}
	if m == 0 {
		return
	}
	for i := range b {
		b[i] *= peak / m
	}
}

// synth composes the effect's signal.
func synth(e Effect) buffer {
	var b buffer
	switch e {
	case Move:
		b = newBuffer(0.12)
		b.knock(0, 620, 1)
	case Capture:
		b = newBuffer(0.22)
		b.tone(0, 330, 0.03, 1)
		b.tone(0, 910, 0.012, 0.5)
		b.click(0, 0.006, 0.9)
		b.knock(0.045, 560, 0.5)
	case Castle:
		b = newBuffer(0.24)
		b.knock(0, 620, 1)
		b.knock(0.11, 540, 0.9)
	case Check:
		b = newBuffer(0.36)
		b.knock(0, 620, 0.8)
		b.bell(0.03, 988, 0.07, 0.6)
		b.bell(0.13, 1319, 0.09, 0.6)
	case Promote:
		b = newBuffer(0.55)
		b.knock(0, 620, 0.8)
		for i, f := range []float64{523.25, 659.25, 783.99, 1046.5} {
			b.bell(0.04+float64(i)*0.07, f, 0.12, 0.5)
		}
	case GameEnd:
		b = newBuffer(1.4)
		for i, f := range []float64{392, 523.25, 659.25} {
			b.bell(float64(i)*0.09, f, 0.45, 0.6)
		}
	case LowTime:
		b = newBuffer(0.3)
		b.bell(0, 1046.5, 0.035, 1)
		b.bell(0.15, 1046.5, 0.035, 1)
	case Error:
		b = newBuffer(0.26)
		for i, f := range []float64{220, 165} {
			at := float64(i) * 0.11
			b.tone(at, f, 0.05, 1)
			b.tone(at, 3*f, 0.04, 0.3) // odd harmonic: a buzzier, "wrong" timbre
		}
	default:
		return newBuffer(0)
	}
	b.normalize()
	return b
}

// WAV returns the effect as a 16-bit mono PCM WAV file.
func WAV(e Effect) []byte {
	b := synth(e)
	const (
		bits     = 16
		channels = 1
	)
	dataLen := uint32(len(b) * bits / 8)
	var out bytes.Buffer
	out.Grow(44 + int(dataLen))
	w := func(v any) { _ = binary.Write(&out, binary.LittleEndian, v) } // bytes.Buffer never fails
	out.WriteString("RIFF")
	w(36 + dataLen)
	out.WriteString("WAVEfmt ")
	w(uint32(16)) // fmt chunk size
	w(uint16(1))  // PCM
	w(uint16(channels))
	w(uint32(SampleRate))
	w(uint32(SampleRate * channels * bits / 8)) // byte rate
	w(uint16(channels * bits / 8))              // block align
	w(uint16(bits))
	out.WriteString("data")
	w(dataLen)
	for _, v := range b {
		w(int16(math.Round(max(-1, min(1, v)) * math.MaxInt16)))
	}
	return out.Bytes()
}
