# polly

A small Go CLI that helps you study English vocabulary and idioms.

Given a word or phrase as an argument, it:

1. Uses the [GitHub Copilot SDK](https://github.com/github/copilot-sdk) to generate an example sentence, a Japanese translation, and a short explanation,
2. Generates an `.mp3` file of the example sentence read aloud by AWS Polly, and
3. Saves the formatted example sentence, translation, and explanation to the clipboard (via macOS `pbcopy`).

## Usage

```bash
polly "eager"
polly "look forward to"
```

Example output:

```
Audio file created: She_is_eager_to_learn_new_languages.mp3
Copied vocab card to clipboard.
```

The clipboard will contain text like the following:

```
- She is eager to learn new languages.
- 彼女は新しい言語を学ぶことに熱心だ。

- 「eager」は「熱心な」「強く望んでいる」という意味の形容詞。
```

The `.mp3` filename is a sanitized version of the example sentence: whitespace, punctuation, and symbols are replaced with `_`, consecutive `_` characters are collapsed to a single one, and leading/trailing `_` characters are removed.

## Prerequisites

- The [Copilot CLI](https://github.com/features/copilot/cli) must be installed and the `copilot` command must be on your `PATH` (already authenticated via `gh auth login`). The SDK invokes the Copilot CLI at startup.
- The AWS CLI must be installed and configured with credentials that can access AWS Polly.
- macOS is required (the clipboard integration uses `pbcopy`).

## Development

```bash
go test ./...
go build .
```

## Running a released binary

Download the `polly-<os>-<arch>` binary (with `.exe` on Windows) that matches your OS and architecture from [GitHub Releases](https://github.com/mahata/polly/releases).

### macOS

```bash
chmod +x ./polly-darwin-arm64
xattr -d com.apple.quarantine ./polly-darwin-arm64
./polly-darwin-arm64 "eager"
```

The same steps work for the `amd64` build. Alternatively, you can approve the binary in Finder by right-clicking and choosing "Open" → "Open".
