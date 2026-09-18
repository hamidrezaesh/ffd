# Usage

## Root

### Basic

```bash
ffd <URL>
```

Download multiple files:

```bash
ffd <URL1> <URL2> <URL3>
```

### Options

| Option                 | Short | Description                 | Default |
| ---------------------- | ----- | --------------------------- | ------- |
| `--output NAME`        | `-o`  | Custom output filename      | —       |
| `--wait SECONDS`       | `-w`  | Wait before downloading     | —       |
| `--path PATH`          | `-p`  | Output directory            | `.`     |
| `--max-retries NUMBER` | `-r`  | Maximum retries             | `4`     |
| `--max-workers NUMBER` | `-W`  | Maximum concurrent workers  | `8`     |
| `--max-chunks NUMBER`  | `-c`  | Maximum download chunks     | `12`    |
| `--protocol PROTOCOL`  | —     | HTTP protocol to use        | `auto`  |
| `--set-proxy PROXY`    | —     | Set a proxy for downloading | —       |
| `--help`               | `-h`  | Show help                   | —       |
| `--version`            | `-v`  | Show version                | —       |

### Examples

Custom filename:

```bash
ffd <URL> -o my-file.zip
```

Custom download directory:

```bash
ffd <URL> -p ~/Downloads
```

More workers and chunks:

```bash
ffd <URL> -W 16 -c 20
```

Force HTTP/2:

```bash
ffd <URL> --protocol http2
```

Wait before downloading:

```bash
ffd <URL> -w 10
```

Increase retries:

```bash
ffd <URL> -r 10
```

Set a proxy for downloading:

```bash
ffd <URL> --set-proxy http://127.0.0.1:8000
```

---

## Proxy

Start the ffd forward proxy:

```bash
ffd proxy
```

The proxy listens on `127.0.0.1:8000` by default.

### Options

| Option              | Description                             | Default |
| ------------------- | --------------------------------------- | ------- |
| `--port PORT`       | Set the local proxy listening port      | `8000`  |
| `--set-proxy PROXY` | Set an upstream proxy for the ffd proxy | —       |
| `--help`            | Show help                               | —       |

### Examples

Set a custom port:

```bash
ffd proxy --port 9000
```

Set an upstream proxy:

```bash
ffd proxy --set-proxy http://127.0.0.1:9000
```

Set both:

```bash
ffd proxy --port 9000 --set-proxy http://127.0.0.1:8080
```

---

## Update

Update ffd to the latest version:

```bash
ffd update
```
