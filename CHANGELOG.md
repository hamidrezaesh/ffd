# Changelog
All notable changes to FFD are documented in this file.

## [Unreleased]
### Added
* Added upstream proxy support.
* Added the --set-proxy option for configuring an upstream proxy.

## Fixed
* Fixed version comparison in `ffd update` when GitHub release tags include the `v` prefix.

## v0.3.11
### Performance
* Added protocol availability checks to avoid attempting unsupported protocols.

## v0.3.1

### Added
* Added CLI version information.
* Added request logging to the proxy.

### Changed
* Improved the CLI user experience and command behavior.

## v0.3.0 - Network Optimization Update

### Added
* Added protocol selection to the CLI.
* Added automatic HTTP protocol testing.
* Added HTTP/3 support.
* Added preferred protocol selection.
* Added dynamic worker selection for improved download performance.

### Changed
* Optimized the HTTP transport layer.
* Improved range splitting and replaced duplicate range-splitting logic with a shared implementation.
* Refactored the scheduler and download pipeline to simplify protocol and worker management.
* Moved the HTTP client from the engine/server layer into the scheduler.

### Fixed
* Fixed automatic protocol selection by correctly using `0` as the auto-selection value.

## v0.2.0

### Added
* Added the `ffd proxy` command.
* Added an accelerated forward proxy for faster request forwarding.
* Added a Windows PowerShell installation script.

### Changed
* Improved chunk processing by streaming chunks through channels.
* Updated the engine to write downloaded chunks more efficiently.
* Improved the installation script to use the repository variable.

## v0.1.13

### Added
* Added an installer script for Unix-based systems.

### Performance
* Improved download scheduling with a worker pool.

## v0.1.12

### Added
* Added automatic retries for failed downloads.

### Changed
* Improved HTTP transport configuration for better download performance.

### Fixed
* Added validation for HTTP response status codes to prevent invalid responses from being processed.


## v0.1.1

### Fixed
* Fixed platform-specific builds by adding the required build tags.
