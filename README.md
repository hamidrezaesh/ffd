<div align="center">

# FFD

**Fast File Downloader**

[![Version](https://img.shields.io/badge/version-v0.1.12-2CA5E0)](https://github.com/hamidrezaesh/ffd)
[![Website](https://img.shields.io/badge/website-ffd--cli.pages.dev-111111)](https://ffd-cli.pages.dev)

</div>

`ffd` is a simple command-line tool for downloading files quickly using multiple HTTP byte ranges when supported by the server.

## Features

* Multi-segment downloads
* Automatic file metadata detection
* Automatic filename detection
* Download progress tracking
* Download speed display
* Estimated time remaining
* Custom filenames
* Custom download paths
* Delayed downloads
* Cross-platform Go implementation
* Automatically distributes download chunks among workers for better performance.

## Installation

### Quick Install
#### Linux/Mac
install the latest release with a single command:
```bash
curl -fsSL https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.sh | sh
```

Verify the installation:
```bash
ffd --help
```

#### Windows

If PowerShell blocks script execution, enable local scripts first:

```bash
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
```

Then install the latest release:

```bash
irm https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.ps1 | iex
```

Verify the installation:

```bash
ffd --help
```

### Download a Release

Download the latest version of ffd from the [Releases page](https://github.com/hamidrezaesh/ffd/releases) or [official website](https://ffd-cli.pages.dev/install#:~:text=Install%20from%20Release).

#### Linux / macOS

Extract the downloaded archive:

```bash
tar -xzf ffd_*.tar.gz
```

Then install the binary:

```bash
sudo install -m 755 ffd /usr/local/bin/ffd
```

Verify the installation:

```bash
ffd --help
```

#### Windows
Download the appropriate .zip archive from the Releases page.
Extract ffd.exe.
Add its directory to your PATH.

Then run:

```bash
ffd --help
```

### From source

Make sure you have Go installed, then clone the repository:

```bash
git clone https://github.com/hamidrezaesh/ffd.git && cd ffd
```

Build the binary:

```bash
go build -o ffd
```

## How it works

When the server supports HTTP byte-range requests, `ffd` divides the file into multiple ranges and downloads them concurrently.

If the server does not support range requests, `ffd` automatically falls back to a single download stream.

## Project Structure

```text
ffd/
├── cmd/
├── internal/
│   ├── disk/
│   ├── engine/
│   ├── formatter/
│   ├── metadata/
│   ├── scheduler/
│   ├── tracker/
│   └── validator/
├── main.go
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

## Requirements

* Go 1.26 or newer for building from source

## License

`ffd` is licensed under the MIT License.

See [LICENSE](./LICENSE) for the full license text.

Made by [Hamidreza](https://github.com/hamidrezaesh).
