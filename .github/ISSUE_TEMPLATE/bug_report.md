---
name: Bug Report
about: Report a bug or unexpected behavior in FFD
title: "[Bug]: "
labels: bug
assignees: ''
---

## Description

A clear and concise description of the issue you're experiencing.

**Example:**

FFD fails to resume interrupted downloads. When I restart a download, it starts from the beginning instead of continuing from where it stopped.

## Target URL

Provide the URL of the file or resource you were trying to download.

**Example:**

`https://example.com/files/archive.zip`

## Steps to Reproduce

Describe the steps required to reproduce the issue.

**Example:**

1. Run `ffd https://example.com/files/archive.zip`
2. Interrupt the download.
3. Run the same command again.
4. Notice that the download starts from the beginning.

## Expected Behavior

Describe what you expected FFD to do.

**Example:**

FFD should detect the existing partial download and resume from where it stopped.

## Actual Behavior

Describe what actually happened instead.

**Example:**

FFD starts downloading the entire file again.

## Environment

Provide information about your environment to help us investigate the issue.

- FFD version: `0.4.0-beta.2`
- Operating system: `Ubuntu 24.04`
- Architecture: `amd64`
- Installation method: `Official Installer Script`

## Logs / Error Output

Include relevant logs or error messages, if available.

**Example:**

```text
Error: failed to resume download
server returned unexpected status code: 200
```

## Additional Context

Provide any additional information that might help us understand or reproduce the issue.

**Example:**

The issue only occurs with a specific server. Other download URLs work as expected.

---

**Note:** Remove any sensitive information, such as access tokens, API keys, or personal credentials, before sharing URLs or logs.
