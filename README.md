# polly

英単語・英熟語の学習を支援する小さな Go CLI。

引数で渡した語/熟語に対して、

1. [GitHub Copilot SDK](https://github.com/github/copilot-sdk) で例文・和訳・短い解説を生成し、
2. 例文を AWS Polly で読み上げた `.mp3` ファイルを生成し、
3. 例文・和訳・解説を整形した状態でクリップボード (macOS `pbcopy`) に保存します。

## Usage

```bash
polly "eager"
polly "look forward to"
```

出力例:

```
Audio file created: She_is_eager_to_learn_new_languages.mp3
Copied vocab card to clipboard.
```

クリップボードには次のようなテキストが入ります。

```
- She is eager to learn new languages.
- 彼女は新しい言語を学ぶことに熱心だ。

- 「eager」は「熱心な」「強く望んでいる」という意味の形容詞。
```

`.mp3` のファイル名は例文を sanitize したもので、空白・句読点・記号は `_` に置換し、連続する `_` は 1 個に圧縮、先頭/末尾の `_` は削除します。

## 前提

- [Copilot CLI](https://github.com/features/copilot/cli) がインストール済みで `copilot` コマンドが PATH にあること (`gh auth login` でログイン済み)。SDK が起動時に Copilot CLI を呼び出します。
- AWS CLI がインストール済みで、AWS Polly にアクセスできる資格情報が設定済みであること。
- macOS であること (クリップボードに `pbcopy` を使用)。

## Development

```bash
go test ./...
go build .
```

## Running a released binary

OS/アーキテクチャに合った `polly-<os>-<arch>` (Windows は `.exe`) を [GitHub Releases](https://github.com/mahata/polly/releases) からダウンロードしてください。

### macOS

```bash
chmod +x ./polly-darwin-arm64
xattr -d com.apple.quarantine ./polly-darwin-arm64
./polly-darwin-arm64 "eager"
```

`amd64` ビルドも同じ手順です。Finder で右クリック → "開く" → "開く" で承認することもできます。
