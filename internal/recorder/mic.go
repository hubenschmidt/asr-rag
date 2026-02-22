package recorder

import (
	"fmt"
	"os/exec"
	"strconv"
)

// Record captures audio from the default mic for the given number of seconds.
// Writes a WAV file (16kHz, mono, 16-bit PCM) to outPath using arecord (ALSA).
func Record(seconds int, outPath string) error {
	fmt.Printf("recording %d seconds...\n", seconds)

	cmd := exec.Command("arecord",
		"-f", "S16_LE",
		"-r", "16000",
		"-c", "1",
		"-d", strconv.Itoa(seconds),
		outPath,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("arecord: %w", err)
	}

	fmt.Println("recording done")
	return nil
}
