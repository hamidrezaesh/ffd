# FFD

![Version](https://img.shields.io/badge/version-v0.1.12-2CA5E0)

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

## Installation

### Download a Release

Download the latest version of ffd from the Releases page.

Choose the archive matching your operating system and CPU architecture:

| Operating System | Architecture |
|------------------|---------------|
| Linux | [x86_64/amd64](https://github.com/hamidrezaesh/ffd/releases/download/v0.1.12/ffd_0.1.12_linux_amd64.tar.gz) |
| Linux | [ARM64](https://github.com/hamidrezaesh/ffd/releases/download/v0.1.1/ffd_0.12.12_linux_arm64.tar.gz) |
| macOS | [x86_64/amd64](https://github.com/hamidrezaesh/ffd/releases/download/v0.1.12/ffd_0.1.12_darwin_amd64.tar.gz) |
| macOS | [ARM64](https://github.com/hamidrezaesh/ffd/releases/download/v0.1.12/ffd_0.1.12_darwin_arm64.tar.gz) |
| Windows | [x86_64/amd64](https://github.com/hamidrezaesh/ffd/releases/download/v0.1.12/ffd_0.1.12_windows_amd64.zip) |
| Windows | [ARM64](https://github.com/hamidrezaesh/ffd/releases/download/v0.1.12/ffd_0.1.12_windows_arm64.zip) |

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
