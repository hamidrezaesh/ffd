

<div align="center">

# FFD

**Fast File Downloader**

[v0.3.13](https://github.com/hamidrezaesh/ffd/releases)|
[Website](https://ffd-cli.pages.dev)|
[Issue](https://github.com/hamidrezaesh/ffd/issues)

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
* Dynamic worker selection while downloading.
* Built-in proxy — Use FFD as a local HTTP forward proxy.

## Installation

**Linux / macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.sh | sh
```

**Windows**

```powershell
irm https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.ps1 | iex
```

For detailed instructions and alternative installation methods, see **[docs/INSTALLATION.md](docs/INSTALLATION.md)**.


## Usage
### Download a file
```bash
ffd https://example.com/file.zip
```

ffd automatically detects the file metadata and, when supported, downloads the file using multiple HTTP byte ranges.

### HTTP Forward Proxy

ffd can also run as a local HTTP forward proxy.

Start the proxy:

```bash
ffd proxy
```

By default, the proxy listens on `127.0.0.1:8000`

You can change the port with:

```bash
ffd proxy --port 9000
```

Then configure your browser or another HTTP client to use:

HTTP Proxy: 127.0.0.1:8000  
HTTPS Proxy: 127.0.0.1:8000

For detailed information about using ffd, including commands, options, and examples, see **[docs/USAGE.md](docs/USAGE.md)**.

## How it works

When the server supports HTTP byte-range requests, `ffd` divides the file into multiple ranges and downloads them concurrently.

If the server does not support range requests, `ffd` automatically falls back to a single download stream.

## Project Structure

```text
ffd/
├── .github/
|   ├── workflows/
├── cmd/
├── internal/
│   ├── disk/
│   ├── engine/
│   ├── formatter/
│   ├── metadata/
│   ├── proxy/
│   ├── scheduler/
│   ├── tracker/
│   └── validator/
├── scripts/
├── docs/
├── .gitignore
├── .goreleaser.yaml
├── main.go
├── CONTRIBUTING.md
├── CHANGELOG.md
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

For detailed information about the project structure, packages, files, and how ffd works internally, see **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)**.

## Requirements

* Go 1.26 or newer for building from source

## License

`ffd` is licensed under the Apache 2.0 License.

See [LICENSE](./LICENSE) for the full license text.

Made by [Hamidreza](https://github.com/hamidrezaesh).
