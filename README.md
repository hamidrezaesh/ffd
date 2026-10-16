

<div align="center">

# FFD

**Fast File Downloader**

[v0.4.0](https://github.com/hamidrezaesh/ffd/releases)|
[Website](https://ffd-cli.pages.dev)|
[Issue](https://github.com/hamidrezaesh/ffd/issues)

</div>

`ffd` is a simple command-line tool for downloading files quickly using multiple HTTP byte ranges when supported by the server.

## Features

* Faster downloads with multi-segment downloading.
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

### HTTP Forward Proxy

```bash
ffd proxy
```

## How it works

When the server supports HTTP byte-range requests, `ffd` divides the file into multiple ranges and downloads them concurrently.

If the server does not support range requests, `ffd` automatically falls back to a single download stream.

## Project Structure

```text
ffd/
├── .github/
│   └── workflows/
├── cmd/
├── internal/
│   ├── config/
│   ├── disk/
│   ├── engine/
│   ├── formatter/
│   ├── metadata/
│   ├── proxy/
│   ├── scheduler/
│   ├── tracker/
│   └── tools/
│   └── state/
├── scripts/
├── version/
├── docs/
├── test/
├── .gitignore
├── .goreleaser.yaml
├── main.go
├── go.mod
├── CONTRIBUTING.md
├── go.sum
├── README.md
└── LICENSE
```

For detailed information about the project structure, packages, files, and how ffd works internally, see **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)**.

## Contributing
Want to contribute to FFD? Check out the [Contributing Guide](CONTRIBUTING.md) to get started.

## Roadmap
Check out [Roadmap](https://ffd-cli.pages.dev/roadmap) to see planned features and
future improvements.
> Have an idea? Read [`Contributing.md`](CONTRIBUTING.md#proposing-roadmap-ideas) for more details.

## License

`ffd` is licensed under the Apache 2.0 License.

See [LICENSE](./LICENSE) for the full license text.

Made by [Hamidreza](https://github.com/hamidrezaesh).
