# Quick Start

This page helps you run the converter in minutes using the CLI or the web app.

## Choose your path

- CLI: fastest and automation-friendly (Go or Rust binaries).
- PowerShell: Windows users who prefer scripts.
- Web app: upload files in a browser and download a zip of CSVs.

## 1) Download a CLI binary

Go to Releases:
https://github.com/IcySparkle/UsxToCsv/releases

Pick one:
- `usxtocsv-go-...` (Go build)
- `usxtocsv-rust-...` (Rust build)

Extract the archive to a folder on your machine.

## 2) Run a single file

```bash
./usxtocsv -input "/path/to/FILE.usx"
```

## 3) Run multiple files

Wildcard:
```bash
./usxtocsv -input "/path/to/*.usx"
```

Multiple inputs:
```bash
./usxtocsv -input "/path/to/MAT.usx" -input "/path/to/MRK.usfm"
```

Output folder:
```bash
./usxtocsv -input "/path/to/*.usx" -output "/path/to/csv"
```

## 4) PowerShell (Windows)

```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\Sources"
```

With output folder:
```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\Sources" -OutputFolder "C:\Bible\CSV"
```

## 5) Web App (Docker)

```bash
docker run -p 8080:8080 ghcr.io/icysparkle/usxtocsv-web:latest
```

Open:
- React UI: `http://localhost:8080`
- Simple UI: `http://localhost:8080/simple`

Upload `.usx`, `.usfm`, `.sfm`, or a `.zip` containing them.

## 6) Automation (optional)

JSON summary:
```bash
./usxtocsv -input "/path/to/FILE.usx" -json
```

Quiet mode:
```bash
./usxtocsv -input "/path/to/FILE.usx" -quiet
```

## Troubleshooting quick tips

- "No files found": check the path or wildcard.
- "Input must be a .usx, .usfm, or .sfm file": check file extensions.
- Web app errors: confirm uploads are supported types.
