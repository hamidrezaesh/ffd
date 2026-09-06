# Project Structure

The ffd source code is organized into separate packages, with each package responsible for a specific part of the downloader.

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
├── .gitignore
├── .goreleaser.yaml
├── main.go
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

## `main.go`

The entry point of ffd.

It starts the Cobra CLI by calling the root command.

---

## `cmd/`

Contains the CLI commands and their configuration.

### `cmd/root.go`

Defines the main `ffd` command.

It handles:

* CLI arguments
* download flags
* multiple URLs
* protocol selection
* workers and chunks
* retries
* output path

### `cmd/version.go`

Contains the build-time version variable used by `ffd --version`.

---

# `internal/`

Contains ffd's internal implementation.

These packages are not intended to be imported by external Go projects.

## `internal/engine/`

Responsible for the actual download process.

It handles:

* HTTP requests
* downloading file ranges
* workers
* writing downloaded data
* retries
* HTTP client configuration

The engine receives the download plan from the scheduler and executes it.

---

## `internal/scheduler/`

Responsible for deciding **how the download should be performed**.

It handles things such as:

* splitting files into chunks
* selecting the number of workers
* testing HTTP protocols
* selecting the preferred protocol
* creating download ranges

The scheduler produces the work that the engine executes.

```text
URL
 │
 ▼
Scheduler
 │
 ├── Protocol
 ├── Workers
 └── Chunks
 │
 ▼
Engine
 │
 ▼
Downloaded File
```

---

## `internal/proxy/`

Contains the ffd forward proxy.

It handles:

* incoming proxy requests
* HTTP requests
* HTTPS `CONNECT` tunnels
* checking whether a resource supports ranged downloads
* forwarding requests
* accelerated downloads when possible

---

# `scripts/`

Contains installation scripts used by the updater.

### `scripts/install.sh`

Installation script for Linux and macOS.

### `scripts/install.ps1`

Installation script for Windows PowerShell.

---

# Configuration & Build Files

## `go.mod`

Defines the Go module and project dependencies.

## `go.sum`

Contains checksums for Go dependencies.

## `.goreleaser.yaml`

Configuration for building and releasing ffd.

It defines:

* supported operating systems
* supported architectures
* binary name
* archives
* checksums
* changelog generation
* version injection

# `.github/workflows/`

Contains GitHub Actions workflows used to automatically build, test, and release ffd.

### `release.yml`

Runs when a new version tag is pushed, such as:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow:

1. Checks out the repository
2. Sets up Go
3. Runs GoReleaser
4. Builds ffd for supported platforms
5. Creates archives and checksums
6. Publishes the release on GitHub

The workflow allows releases to be created automatically without manually building binaries for every platform.