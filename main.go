package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	copilot "github.com/github/copilot-sdk/go"
)

type vocabResult struct {
	Sentence    string `json:"sentence"`
	Translation string `json:"translation"`
	Explanation string `json:"explanation"`
}

type (
	generateFunc  func(ctx context.Context, term string) (vocabResult, error)
	synthFunc     func(text, filename string) error
	clipboardFunc func(text string) error
)

type deps struct {
	Generate  generateFunc
	Synth     synthFunc
	Clipboard clipboardFunc
}

var (
	underscoreRunRe = regexp.MustCompile(`_+`)
	jsonObjectRe    = regexp.MustCompile(`(?s)\{.*\}`)
)

func sanitizeFilename(text string) string {
	var b strings.Builder
	for _, r := range text {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	collapsed := underscoreRunRe.ReplaceAllString(b.String(), "_")
	trimmed := strings.Trim(collapsed, "_")
	if trimmed == "" {
		trimmed = "sample"
	}
	return trimmed + ".mp3"
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

func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

const vocabPromptTemplate = `You are a helpful English tutor for a Japanese learner.

Given the English word or phrase: %q

Do all of the following:
1. Write ONE natural English example sentence (about 8 to 20 words) that contains the given word or phrase exactly as written (case-insensitive match is OK).
2. Provide a natural Japanese translation of that sentence.
3. Provide a short Japanese explanation (1 to 3 sentences) of the meaning and typical usage of the given word or phrase.

Respond with ONLY a single JSON object on one line, with no surrounding prose, no markdown, no code fences, exactly these keys:
{"sentence": "<the English example sentence>", "translation": "<Japanese translation>", "explanation": "<Japanese explanation>"}`

func parseVocabJSON(raw string) (vocabResult, error) {
	trimmed := strings.TrimSpace(raw)

	// Strip ```json ... ``` or ``` ... ``` fences.
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimPrefix(trimmed, "json")
		trimmed = strings.TrimPrefix(trimmed, "JSON")
		if idx := strings.LastIndex(trimmed, "```"); idx >= 0 {
			trimmed = trimmed[:idx]
		}
		trimmed = strings.TrimSpace(trimmed)
	}

	var v vocabResult
	if err := json.Unmarshal([]byte(trimmed), &v); err == nil && v.Sentence != "" {
		return v, nil
	}

	// Fallback: extract the first {...} block from the raw text.
	if match := jsonObjectRe.FindString(raw); match != "" {
		if err := json.Unmarshal([]byte(match), &v); err == nil && v.Sentence != "" {
			return v, nil
		}
	}

	return vocabResult{}, fmt.Errorf("failed to parse vocab JSON from response: %s", raw)
}

func generateVocabContent(ctx context.Context, term string) (vocabResult, error) {
	client := copilot.NewClient(&copilot.ClientOptions{LogLevel: "error"})
	if err := client.Start(ctx); err != nil {
		return vocabResult{}, fmt.Errorf("start copilot client: %w", err)
	}
	defer client.Stop()

	session, err := client.CreateSession(ctx, &copilot.SessionConfig{
		OnPermissionRequest: copilot.PermissionHandler.ApproveAll,
	})
	if err != nil {
		return vocabResult{}, fmt.Errorf("create session: %w", err)
	}
	defer session.Disconnect()

	reply, err := session.SendAndWait(ctx, copilot.MessageOptions{
		Prompt: fmt.Sprintf(vocabPromptTemplate, term),
	})
	if err != nil {
		return vocabResult{}, fmt.Errorf("send message: %w", err)
	}
	if reply == nil {
		return vocabResult{}, errors.New("empty reply from copilot")
	}
	msg, ok := reply.Data.(*copilot.AssistantMessageData)
	if !ok || msg == nil {
		return vocabResult{}, errors.New("unexpected reply type from copilot")
	}

	return parseVocabJSON(msg.Content)
}

func formatClipboard(v vocabResult) string {
	return fmt.Sprintf("- %s\n- %s\n\n- %s\n", v.Sentence, v.Translation, v.Explanation)
}

func run(ctx context.Context, args []string, progName string, out io.Writer, d deps) int {
	if len(args) == 0 {
		fmt.Fprintf(out, "Usage: %s \"English word or phrase\"\n", progName)
		return 1
	}

	term := args[0]

	vocab, err := d.Generate(ctx, term)
	if err != nil {
		fmt.Fprintf(out, "Failed to generate vocab content: %v\n", err)
		return 1
	}

	filename := sanitizeFilename(vocab.Sentence)
	if err := d.Synth(vocab.Sentence, filename); err != nil {
		fmt.Fprintln(out, "Failed to generate audio.")
		return 1
	}
	fmt.Fprintf(out, "Audio file created: %s\n", filename)

	clip := formatClipboard(vocab)
	if err := d.Clipboard(clip); err != nil {
		fmt.Fprintf(out, "Failed to copy to clipboard: %v\n", err)
		return 1
	}
	fmt.Fprintln(out, "Copied vocab card to clipboard.")
	return 0
}

func main() {
	exitCode := run(
		context.Background(),
		os.Args[1:],
		os.Args[0],
		os.Stdout,
		deps{
			Generate:  generateVocabContent,
			Synth:     synthesizeSpeech,
			Clipboard: copyToClipboard,
		},
	)
	os.Exit(exitCode)
}
