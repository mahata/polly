# polly

Small Go CLI that synthesizes speech with AWS Polly.

## Usage

```bash
go run . "text to synthesize"
```

## Running a released binary

Download the binary that matches your OS/arch (`polly-<os>-<arch>`, with `.exe` on Windows) from [GitHub Releases](https://github.com/mahata/polly/releases).

You need the AWS CLI installed and credentials configured that can access AWS Polly.

### macOS

The downloaded binary has the quarantine attribute set, so running it as-is will be blocked by Gatekeeper with "Apple could not verify ...". Since the binary is not code-signed or notarized, you need to allow it manually the first time.

```bash
chmod +x ./polly-darwin-arm64
xattr -d com.apple.quarantine ./polly-darwin-arm64
./polly-darwin-arm64 "text to synthesize"
```

The `amd64` build uses the same steps. Alternatively, instead of using the CLI, you can right-click the binary in Finder, choose "Open", and then "Open" again to approve it on first launch.

### Linux

```bash
chmod +x ./polly-linux-amd64
./polly-linux-amd64 "text to synthesize"
```

### Windows

```powershell
.\polly-windows-amd64.exe "text to synthesize"
```

Because the binary is unsigned, SmartScreen may show a warning on first launch. Click "More info" → "Run anyway" to continue.
