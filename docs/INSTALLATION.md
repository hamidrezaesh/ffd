# Installation

There are several ways to install ffd. Choose the method that works best for you.

## Quick Install

The easiest way to install the latest release is using the official installation scripts.

### Linux / macOS

Run:

```bash
curl -fsSL https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.sh | sh
```

Verify the installation:

```bash
ffd --help
```

### Windows

If PowerShell blocks script execution, enable local scripts:

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
```

Then install the latest release:

```powershell
irm https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.ps1 | iex
```

Verify the installation:

```powershell
ffd --help
```

---

## Download a Release

You can manually download a pre-built release from the [GitHub Releases](https://github.com/hamidrezaesh/ffd/releases?utm_source=chatgpt.com) or the [official website](https://ffd-cli.pages.dev/install?utm_source=chatgpt.com).

### Linux / macOS

Extract the archive:

```bash
tar -xzf ffd_*.tar.gz
```

Install the binary:

```bash
sudo install -m 755 ffd /usr/local/bin/ffd
```

Verify:

```bash
ffd --help
```

### Windows

1. Download the appropriate `.zip` archive.
2. Extract `ffd.exe`.
3. Add its directory to your `PATH`.

Then verify:

```powershell
ffd --help
```

---

## From Source

Make sure you have Go installed.

Clone the repository:

```bash
git clone https://github.com/hamidrezaesh/ffd.git
cd ffd
```

Build ffd:

```bash
go build -o ffd
```

You can then run the binary directly:

```bash
./ffd --help
```

### Installing the Binary

On Linux / macOS, you can optionally install it system-wide:

```bash
sudo install -m 755 ffd /usr/local/bin/ffd
```

Then:

```bash
ffd --help
```

---

## Updating

If ffd is already installed, you can update it to the latest release with:

```bash
ffd update
```