# FFD

![Version](https://img.shields.io/badge/version-v0.1.0-2CA5E0)

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
