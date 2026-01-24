# FAQ

## Go vs Rust builds?
Both are feature-equivalent. Pick either. Go builds are often smaller and faster to compile; Rust builds are strict and safe.

## Why tags?
Tags trigger Releases and GHCR image publishing. This keeps binaries and container images aligned with versions.

## Can I convert folders in the CLI?
Yes. Example:
```
-input "/path/to/folder"
```

## Can I upload a folder via the web app?
Upload a zip containing the folder contents.

## Where are the binaries?
GitHub Releases: https://github.com/IcySparkle/UsxToCsv/releases

## Where is the Docker image?
GHCR: `ghcr.io/icysparkle/usxtocsv-web:<tag>` (use `latest` for the newest tag).

## Does the CLI support JSON output?
Yes. Use `-json` to emit a machine-readable summary to stdout.

## Can I suppress logs for automation?
Yes. Use `-quiet` to suppress progress output.

## Do I need PowerShell if I use the Go/Rust CLI?
No. The Go/Rust binaries are standalone.

## Can I add more styles or columns?
Yes, but it requires code changes. See `UsxToCsv.ps1`, `go/convert`, or `rust/src/main.rs`.
