package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

type synthFunc func(text, filename string) error

func sanitizeFilename(text string) string {
	withUnderscores := strings.ReplaceAll(text, " ", "_")
	sanitized := strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) {
			return -1
		}
		return r
	}, withUnderscores)
	return sanitized + ".mp3"
}

func synthesizeSpeech(text, filename string) error {
	cmd := exec.Command(
		"aws",
		"polly",
		"synthesize-speech",
		"--output-format", "mp3",
		"--voice-id", "Joanna",
		"--engine", "neural",
		"--text", text,
		filename,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func run(args []string, progName string, out io.Writer, synth synthFunc) int {
	if len(args) == 0 {
		fmt.Fprintf(out, "Usage: %s \"text to synthesize\"\n", progName)
		return 1
	}

	text := args[0]
	filename := sanitizeFilename(text)

	if err := synth(text, filename); err != nil {
		fmt.Fprintln(out, "Failed to generate audio.")
		return 1
	}

	fmt.Fprintf(out, "Audio file created: %s\n", filename)
	return 0
}

func main() {
	exitCode := run(os.Args[1:], os.Args[0], os.Stdout, synthesizeSpeech)
	os.Exit(exitCode)
}
