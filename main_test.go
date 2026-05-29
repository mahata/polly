package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		{"Hello, world!", "Hello_world.mp3"},
		{"He's eager to ~learn~!", "He_s_eager_to_learn.mp3"},
		{"   spaced   out   ", "spaced_out.mp3"},
		{"multi---dash__test", "multi_dash_test.mp3"},
		{"!!!", "sample.mp3"},
		{"plainword", "plainword.mp3"},
		{"keep digits 123", "keep_digits_123.mp3"},
	}

	for _, tc := range cases {
		got := sanitizeFilename(tc.in)
		if got != tc.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseVocabJSON(t *testing.T) {
	t.Parallel()

	want := vocabResult{
		Sentence:    "She is eager to learn new languages.",
		Translation: "彼女は新しい言語を学ぶことに熱心だ。",
		Explanation: "「eager」は「熱心な」「強く望んでいる」という意味の形容詞。",
	}

	bare := `{"sentence":"She is eager to learn new languages.","translation":"彼女は新しい言語を学ぶことに熱心だ。","explanation":"「eager」は「熱心な」「強く望んでいる」という意味の形容詞。"}`
	fenced := "```json\n" + bare + "\n```"
	withPrelude := "Here is the result:\n" + bare + "\nHope that helps!"

	for name, raw := range map[string]string{
		"bare":         bare,
		"fenced":       fenced,
		"with_prelude": withPrelude,
	} {
		t.Run(name, func(t *testing.T) {
			got, err := parseVocabJSON(raw)
			if err != nil {
				t.Fatalf("parseVocabJSON(%s) error: %v", name, err)
			}
			if got != want {
				t.Fatalf("parseVocabJSON(%s) = %+v, want %+v", name, got, want)
			}
		})
	}
}

func TestParseVocabJSONInvalid(t *testing.T) {
	t.Parallel()

	if _, err := parseVocabJSON("not json at all"); err == nil {
		t.Fatal("expected error for non-JSON input")
	}
	if _, err := parseVocabJSON(`{"sentence":""}`); err == nil {
		t.Fatal("expected error when sentence is empty")
	}
}

func TestFormatClipboard(t *testing.T) {
	t.Parallel()

	v := vocabResult{
		Sentence:    "She is eager to learn.",
		Translation: "彼女は学ぶことに熱心だ。",
		Explanation: "「eager」は熱心な、強く望むという意味。",
	}
	got := formatClipboard(v)
	want := "- She is eager to learn.\n- 彼女は学ぶことに熱心だ。\n\n- 「eager」は熱心な、強く望むという意味。\n"
	if got != want {
		t.Fatalf("formatClipboard = %q, want %q", got, want)
	}
}

type fakeRun struct {
	genCalls     int
	synthCalls   int
	clipCalls    int
	gotSynthText string
	gotSynthFile string
	gotClipText  string
	genResult    vocabResult
	genErr       error
	synthErr     error
	clipErr      error
}

func (f *fakeRun) deps() deps {
	return deps{
		Generate: func(ctx context.Context, term string) (vocabResult, error) {
			f.genCalls++
			return f.genResult, f.genErr
		},
		Synth: func(text, filename string) error {
			f.synthCalls++
			f.gotSynthText = text
			f.gotSynthFile = filename
			return f.synthErr
		},
		Clipboard: func(text string) error {
			f.clipCalls++
			f.gotClipText = text
			return f.clipErr
		},
	}
}

func TestRunNoArgs(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	f := &fakeRun{}
	code := run(context.Background(), nil, "polly", &out, f.deps())

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if f.genCalls+f.synthCalls+f.clipCalls != 0 {
		t.Fatalf("unexpected calls: %+v", f)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("output missing usage: %q", out.String())
	}
}

func TestRunSuccess(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	f := &fakeRun{
		genResult: vocabResult{
			Sentence:    "She is eager to learn new languages.",
			Translation: "彼女は新しい言語を学ぶことに熱心だ。",
			Explanation: "「eager」は熱心な、強く望むという意味。",
		},
	}
	code := run(context.Background(), []string{"eager"}, "polly", &out, f.deps())

	if code != 0 {
		t.Fatalf("exit = %d, want 0, out=%q", code, out.String())
	}
	if f.genCalls != 1 || f.synthCalls != 1 || f.clipCalls != 1 {
		t.Fatalf("call counts wrong: %+v", f)
	}
	if f.gotSynthText != f.genResult.Sentence {
		t.Fatalf("synth text = %q, want %q", f.gotSynthText, f.genResult.Sentence)
	}
	wantFile := "She_is_eager_to_learn_new_languages.mp3"
	if f.gotSynthFile != wantFile {
		t.Fatalf("synth file = %q, want %q", f.gotSynthFile, wantFile)
	}
	wantClip := "- She is eager to learn new languages.\n- 彼女は新しい言語を学ぶことに熱心だ。\n\n- 「eager」は熱心な、強く望むという意味。\n"
	if f.gotClipText != wantClip {
		t.Fatalf("clip text = %q, want %q", f.gotClipText, wantClip)
	}
	if !strings.Contains(out.String(), wantFile) {
		t.Fatalf("output missing filename: %q", out.String())
	}
	if !strings.Contains(out.String(), "clipboard") {
		t.Fatalf("output missing clipboard confirmation: %q", out.String())
	}
}

func TestRunGenerateFailure(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	f := &fakeRun{genErr: errors.New("boom")}
	code := run(context.Background(), []string{"eager"}, "polly", &out, f.deps())

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if f.synthCalls != 0 || f.clipCalls != 0 {
		t.Fatalf("should not synth/clip on generate failure: %+v", f)
	}
	if !strings.Contains(out.String(), "boom") {
		t.Fatalf("output missing error: %q", out.String())
	}
}

func TestRunSynthFailure(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	f := &fakeRun{
		genResult: vocabResult{Sentence: "Hi there.", Translation: "やあ。", Explanation: "挨拶。"},
		synthErr:  errors.New("aws failed"),
	}
	code := run(context.Background(), []string{"hi"}, "polly", &out, f.deps())

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if f.clipCalls != 0 {
		t.Fatalf("should not clip on synth failure: %+v", f)
	}
	if !strings.Contains(out.String(), "Failed to generate audio") {
		t.Fatalf("output missing audio failure msg: %q", out.String())
	}
}

func TestRunClipboardFailure(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	f := &fakeRun{
		genResult: vocabResult{Sentence: "Hi there.", Translation: "やあ。", Explanation: "挨拶。"},
		clipErr:   errors.New("pbcopy missing"),
	}
	code := run(context.Background(), []string{"hi"}, "polly", &out, f.deps())

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(out.String(), "Failed to copy to clipboard") {
		t.Fatalf("output missing clipboard failure msg: %q", out.String())
	}
}
