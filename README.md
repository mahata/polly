# polly

Small Go CLI that synthesizes speech with AWS Polly.

## Usage

```bash
go run . "text to synthesize"
```

## Running a released binary

[GitHub Releases](https://github.com/mahata/polly/releases) から自分の OS/arch に合うバイナリ（`polly-<os>-<arch>`、Windows は `.exe`）をダウンロードしてください。

### macOS

ダウンロードしたバイナリには quarantine 属性が付与されており、そのまま実行すると Gatekeeper に "Apple could not verify ..." と表示されてブロックされます。コード署名/公証を行っていないため、初回のみ手動で許可してください。

```bash
chmod +x ./polly-darwin-arm64
xattr -d com.apple.quarantine ./polly-darwin-arm64
./polly-darwin-arm64 "text to synthesize"
```

`amd64` 版も同様の手順です。CLI を使わず Finder で右クリック →「開く」→「開く」を選択しても初回承認できます。

### Linux

```bash
chmod +x ./polly-linux-amd64
./polly-linux-amd64 "text to synthesize"
```

### Windows

```powershell
.\polly-windows-amd64.exe "text to synthesize"
```

署名されていないため、初回起動時に SmartScreen 警告が表示される場合があります。「詳細情報」→「実行」で続行してください。
