package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()

	got := sanitizeFilename("Hello, world!")
	want := "Helloworld.mp3"
	if got != want {
		t.Fatalf("sanitizeFilename() = %q, want %q", got, want)
	}
}

func TestRunNoArgs(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	called := false
	exitCode := run(nil, "polly", &out, func(string, string) error {
		called = true
		return nil
	})

	if exitCode != 1 {
		t.Fatalf("run() exit code = %d, want 1", exitCode)
	}
	if called {
		t.Fatal("synth function was called unexpectedly")
	}
	if got, want := out.String(), "Usage: polly \"text to synthesize\"\n"; got != want {
		t.Fatalf("run() output = %q, want %q", got, want)
	}
}

func TestRunSuccess(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var gotText, gotFilename string
	exitCode := run([]string{"Hello, world!"}, "polly", &out, func(text, filename string) error {
		gotText = text
		gotFilename = filename
		return nil
	})

	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0", exitCode)
	}
	if gotText != "Hello, world!" {
		t.Fatalf("synth text = %q, want %q", gotText, "Hello, world!")
	}
	if gotFilename != "Helloworld.mp3" {
		t.Fatalf("synth filename = %q, want %q", gotFilename, "Helloworld.mp3")
	}
	if got, want := out.String(), "Audio file created: Helloworld.mp3\n"; got != want {
		t.Fatalf("run() output = %q, want %q", got, want)
	}
}

func TestRunFailure(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	exitCode := run([]string{"Hello"}, "polly", &out, func(string, string) error {
		return errors.New("aws failed")
	})

	if exitCode != 1 {
		t.Fatalf("run() exit code = %d, want 1", exitCode)
	}
	if got, want := out.String(), "Failed to generate audio.\n"; got != want {
		t.Fatalf("run() output = %q, want %q", got, want)
	}
}
