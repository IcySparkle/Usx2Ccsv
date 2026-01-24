# CLI Usage

This page covers the Go and Rust CLIs in depth, plus the PowerShell script for Windows users.

## Go CLI

### Build from source
```bash
cd go
go build -o usxtocsv .
```

### Basic usage
```bash
./usxtocsv -input "/path/to/FILE.usx"
./usxtocsv -input "/path/to/FILE.usfm"
./usxtocsv -input "/path/to/FILE.sfm"
```

### Folder input
```bash
./usxtocsv -input "/path/to/folder"
```

### Wildcards
```bash
./usxtocsv -input "/path/to/*.usx"
```

### Multiple inputs
```bash
./usxtocsv -input "/path/to/MAT.usx" -input "/path/to/MRK.usfm"
```

### Output folder
```bash
./usxtocsv -input "/path/to/*.usx" -output "/path/to/csv"
```

### Automation output
```bash
./usxtocsv -input "/path/to/FILE.usx" -json
./usxtocsv -input "/path/to/FILE.usx" -quiet
```

### Help
```bash
./usxtocsv -help
```

## Rust CLI

### Build from source
```bash
cd rust
cargo build --release
```

### Basic usage
```bash
./target/release/usxtocsv -input "/path/to/FILE.usx"
./target/release/usxtocsv -input "/path/to/FILE.usfm"
./target/release/usxtocsv -input "/path/to/FILE.sfm"
```

### Folder input
```bash
./target/release/usxtocsv -input "/path/to/folder"
```

### Wildcards
```bash
./target/release/usxtocsv -input "/path/to/*.usx"
```

### Multiple inputs
```bash
./target/release/usxtocsv -input "/path/to/MAT.usx" -input "/path/to/MRK.usfm"
```

### Output folder
```bash
./target/release/usxtocsv -input "/path/to/*.usx" -output "/path/to/csv"
```

### Automation output
```bash
./target/release/usxtocsv -input "/path/to/FILE.usx" -json
./target/release/usxtocsv -input "/path/to/FILE.usx" -quiet
```

### Help
```bash
./target/release/usxtocsv -help
```

## PowerShell (Windows)

### Single file
```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\JHN.usx"
```

### Folder input
```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\Sources"
```

### Wildcards
```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\*.usx"
```

### Multiple inputs
```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\MAT.usx","C:\Bible\MRK.usfm"
```

### Output folder
```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\Sources" -OutputFolder "C:\Bible\CSV"
```

### Help
```powershell
.\UsxToCsv.ps1 -Help
```

## Exit codes
- 0 = success
- 1 = error

## Tips
- Quotes are recommended for paths with spaces.
- The web app accepts `.zip` uploads but the CLI expects files/folders.
- `-json` outputs machine-readable results to stdout; progress logs go to stderr.
