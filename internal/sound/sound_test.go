package sound

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWAVHeaders(t *testing.T) {
	for _, e := range Effects() {
		data := WAV(e)
		if string(data[0:4]) != "RIFF" || string(data[8:16]) != "WAVEfmt " || string(data[36:40]) != "data" {
			t.Fatalf("%v: malformed header", e)
		}
		if got := binary.LittleEndian.Uint32(data[4:8]); int(got) != len(data)-8 {
			t.Errorf("%v: RIFF size %d, want %d", e, got, len(data)-8)
		}
		if got := binary.LittleEndian.Uint32(data[24:28]); got != SampleRate {
			t.Errorf("%v: sample rate %d", e, got)
		}
		dataLen := binary.LittleEndian.Uint32(data[40:44])
		if int(dataLen) != len(data)-44 || dataLen == 0 {
			t.Errorf("%v: data size %d for %d bytes", e, dataLen, len(data))
		}
	}
}

func TestEffectsAreNormalizedAndDistinct(t *testing.T) {
	seen := map[string]Effect{}
	for _, e := range Effects() {
		data := WAV(e)
		var loudest int
		for i := 44; i < len(data); i += 2 {
			v := int(int16(binary.LittleEndian.Uint16(data[i:])))
			loudest = max(loudest, v, -v)
		}
		want := int(math.Round(peak * math.MaxInt16))
		if loudest < want-2 || loudest > want+2 {
			t.Errorf("%v: peak %d, want about %d", e, loudest, want)
		}
		if other, dup := seen[string(data)]; dup {
			t.Errorf("%v sounds exactly like %v", e, other)
		}
		seen[string(data)] = e
		if e.String() == "unknown" {
			t.Errorf("effect %d has no name", e)
		}
	}
}

func TestWAVIsDeterministic(t *testing.T) {
	for _, e := range Effects() {
		if !bytes.Equal(WAV(e), WAV(e)) {
			t.Fatalf("%v differs between calls", e)
		}
	}
}

// fakeSystem returns a System that records the commands it would start,
// and a function returning them (waiting for the first one when asked).
func fakeSystem(t *testing.T, goos string, available ...string) (*System, func(wait bool) [][]string) {
	t.Helper()
	s := NewSystem(t.TempDir())
	s.goos = goos
	s.lookPath = func(name string) (string, error) {
		for _, a := range available {
			if a == name {
				return "/usr/bin/" + name, nil
			}
		}
		return "", errors.New("not found")
	}
	var mu sync.Mutex
	var started [][]string
	s.start = func(name string, args ...string) error {
		mu.Lock()
		defer mu.Unlock()
		started = append(started, append([]string{name}, args...))
		return nil
	}
	return s, func(wait bool) [][]string {
		deadline := time.Now().Add(2 * time.Second)
		for wait && time.Now().Before(deadline) {
			mu.Lock()
			n := len(started)
			mu.Unlock()
			if n > 0 {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		mu.Lock()
		defer mu.Unlock()
		return append([][]string(nil), started...)
	}
}

func TestSystemPrefersPipeWire(t *testing.T) {
	s, started := fakeSystem(t, "linux", "aplay", "pw-play", "paplay")
	s.Play(Capture)
	got := started(true)
	if len(got) != 1 || got[0][0] != "/usr/bin/pw-play" || !strings.HasSuffix(got[0][1], "capture-v1.wav") {
		t.Fatalf("started %v", got)
	}
	data, err := os.ReadFile(got[0][1])
	if err != nil || !bytes.Equal(data, WAV(Capture)) {
		t.Fatalf("sound file not written correctly: %v", err)
	}
}

func TestSystemAplayIsQuiet(t *testing.T) {
	s, started := fakeSystem(t, "linux", "aplay")
	s.Play(Move)
	if got := started(true); len(got) != 1 || got[0][0] != "/usr/bin/aplay" || got[0][1] != "-q" {
		t.Fatalf("started %v", got)
	}
}

func TestSystemWindowsUsesPowerShell(t *testing.T) {
	s, started := fakeSystem(t, "windows", "powershell")
	s.Play(Check)
	got := started(true)
	if len(got) != 1 || got[0][0] != "/usr/bin/powershell" {
		t.Fatalf("started %v", got)
	}
	script := got[0][len(got[0])-1]
	if !strings.Contains(script, "Media.SoundPlayer") || !strings.Contains(script, "check-v1.wav") {
		t.Fatalf("script %q", script)
	}
}

func TestSystemWithoutPlayerIsSilent(t *testing.T) {
	s, started := fakeSystem(t, "linux")
	s.Play(Move)
	time.Sleep(50 * time.Millisecond)
	if got := started(false); len(got) != 0 {
		t.Fatalf("started %v without any player", got)
	}
}
