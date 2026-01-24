# UsxToCsv — Convert USX / USFM / SFM Bible Files into a Clean CSV Format

A PowerShell script for converting **USX**, **USFM**, and **SFM** Bible manuscripts into a consistent, publisher-friendly CSV format suitable for:

- Bible layout in InDesign  
- Linguistic / translation analysis  
- Footnote & cross-reference extraction  
- QA automation  
- Parallel Scripture comparison  

The script is fully harmonized across formats: **USX**, **USFM**, and **SFM** produce the same CSV schema and identical verse-level behavior.

---

## 🔥 Features

### ✔ Supports Multiple Formats

| Format | Detection | Notes |
|--------|-----------|-------|
| **USX** (`.usx`) | XML parsing with milestone verse handling | Official USX 3.x compatible |
| **USFM** (`.usfm`) | Token-based parsing | Supports `\v`, `\p`, `\q`, `\qt`, `\+` markers |
| **SFM** (`.sfm`) | Alias for USFM | Fully supported |

---

### ✔ Unified Verse Output Model

Regardless of format:

- One CSV **row per verse**
- Text merged across multiple paragraphs (USX `para`, USFM `\p`/`\m`/`\q`)
- A verse ends when:
  - **USX**: the `</verse eid="">` milestone is encountered  
  - **USFM**: a new `\v` marker appears  

---

### ✔ Inline Formatting → Plain and Styled Output

Inline tags are translated into:

| USX/USFM Style | CSV Output Tag |
|----------------|----------------|
| `wj`   | `<wj>...</wj>` |
| `add`  | `<add>...</add>` |
| `nd`   | `<nd>...</nd>` |
| `it`   | `<i>...</i>` |
| `bd`   | `<b>...</b>` |
| `bdit` | `<bdit>...</bdit>` |
| *(other styles)* | `<span>...</span>` |

CSV provides two columns:

- **TextPlain** → all tags stripped  
- **TextStyled** → GREP-style tags preserved  

Superscript content:

- USX: `<char style="sup">…</char>`  
- USFM: `\sup ... \sup*` and `\+sup ... \+sup*`  

→ **always removed** and not included in either column.

---

### ✔ FT-Only Footnote & Cross-reference Extraction

For both USX and USFM:

- Only **FT (footnote text)** is included  
- Caller, FR (footnote reference), and other meta markers are ignored  
- `Footnotes` and `Crossrefs` columns join multiple items with ` | `  

Examples:

- USX:  
  ```xml
  <note style="f">
    <char style="fr">1:3 </char>
    <char style="ft">Some manuscripts say...</char>
  </note>
  ```
- USFM:  
  ```usfm
  \f + \fr 1:3 \ft Some manuscripts say...\f*
  ```

In both cases, only **`Some manuscripts say...`** appears in the CSV.

---

### ✔ Subtitle / Heading Capture

Recognized as subtitles:

- **USX**: `s, s1, s2, s3, sp, ms, mr, mt, mt1, mt2`  
- **USFM/SFM**: `\s`, `\s1`–`\s3`, `\sp`, `\ms`, `\mr`, `\mt`, `\mt1`, `\mt2`  

Behavior:

- Subtitle is **remembered** until replaced by the next one  
- Every verse row receives the current subtitle in the `Subtitle` column  

---

### ✔ Sanity-Focused Normalization

- All whitespace collapsed to single spaces  
- Line breaks in source are irrelevant for final CSV  
- Unknown backslash markers in USFM are removed unless intentionally mapped  
- Rows are sorted by **Book**, **Chapter (numeric)**, **Verse (string)**  

---

## 📦 CSV Columns

| Column      | Description                                      |
|-------------|--------------------------------------------------|
| **Book**    | USX `<book code="">` or USFM `\id` value         |
| **Chapter** | Numeric chapter number                           |
| **Verse**   | Verse number (supports `1`, `1a`, `1b`, etc.)    |
| **TextPlain**  | Verse text with inline styling removed        |
| **TextStyled** | Verse text with inline styles as GREP tags    |
| **Footnotes**  | FT-only footnotes joined with ` | `           |
| **Crossrefs**  | FT-only cross-references joined with ` | `    |
| **Subtitle**   | Last seen heading text                        |

---

## 🚀 Usage

### Convert a Single File

```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\JHN.usx"
.\UsxToCsv.ps1 -InputPath "C:\Bible\JHN.usfm"
.\UsxToCsv.ps1 -InputPath "C:\Bible\JHN.sfm"
```

### Convert an Entire Folder

```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\Sources"
```

Where `Sources` may contain a mix of:

```text
C:\Bible\Sources\
   ├── MAT.usx
   ├── MRK.usfm
   ├── LUK.sfm
   └── JHN.usx
```

### Specify a Custom Output Folder

```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\Sources" -OutputFolder "C:\Bible\CSV"
```

Each file generates a matching `.csv`:

```text
MAT.csv
MRK.csv
LUK.csv
JHN.csv
```

### Convert with Wildcards or Multiple Inputs

```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\*.usx"
.\UsxToCsv.ps1 -InputPath "C:\Bible\MAT.usx","C:\Bible\MRK.usfm"
```

---

## Go and Rust Versions

Both the Go and Rust CLIs implement the same parsing rules and CSV schema as the PowerShell script. Each builds a standalone binary per platform.

### Go

Build:

```powershell
cd go
go build -o usxtocsv .
```

Run:

```powershell
.\usxtocsv -input "C:\Bible\JHN.usx"
.\usxtocsv -input "C:\Bible\*.usx" -output "C:\Bible\CSV"
.\usxtocsv -input "C:\Bible\MAT.usx" -input "C:\Bible\MRK.usfm"
```

Automation:

```powershell
.\usxtocsv -input "C:\Bible\JHN.usx" -json
.\usxtocsv -input "C:\Bible\JHN.usx" -quiet
```

### Rust

Build:

```powershell
cd rust
cargo build --release
```

Run:

```powershell
.\target\release\usxtocsv -input "C:\Bible\JHN.usx"
.\target\release\usxtocsv -input "C:\Bible\*.usx" -output "C:\Bible\CSV"
.\target\release\usxtocsv -input "C:\Bible\MAT.usx" -input "C:\Bible\MRK.usfm"
```

Automation:

```powershell
.\target\release\usxtocsv -input "C:\Bible\JHN.usx" -json
.\target\release\usxtocsv -input "C:\Bible\JHN.usx" -quiet
```

---

## Web App (Upload -> Download)

A minimal web app is included for browser uploads. It accepts `.usx`, `.usfm`, `.sfm`, or `.zip` files and returns a zip of CSVs.

### Run Locally (Go)

```powershell
cd go
go run .\web
```

Open `http://localhost:8080` in your browser.

### Run with Docker

```bash
docker build -t usxtocsv-web .
docker run -p 8080:8080 usxtocsv-web
```

The Docker image serves the React UI by default. To use the minimal built-in HTML instead, open `http://localhost:8080/simple`.

### React UI (Vite)

The React UI lives in `web-ui` and calls the `/convert` endpoint.

Run locally:

```bash
cd web-ui
npm install
npm run dev
```

By default it calls the API on the same origin. To point at a remote backend:

```bash
VITE_API_BASE="https://your-backend.example.com" npm run dev
```

### Hosting Examples

Render:
- Create a new Web Service from this repo
- Environment: Docker
- Expose port `8080`
- Deploy

Fly.io:
- `fly launch` (choose Dockerfile)
- Ensure `PORT=8080` is set
- `fly deploy`

Railway:
- New Project -> Deploy from GitHub
- Select Dockerfile
- Set `PORT=8080` if prompted
- Deploy

DigitalOcean App Platform:
- Create App from GitHub
- Use Dockerfile
- Set HTTP port to `8080`
- Deploy

---

## Releases

Prebuilt Go and Rust executables are published on GitHub Releases. Download the asset for your platform from:

https://github.com/IcySparkle/UsxToCsv/releases

## Wiki Docs

Extended documentation lives under `docs/` (mirrors the GitHub Wiki content):
- `docs/Home.md`
- `docs/Quick-Start.md`
- `docs/CLI-Usage.md`
- `docs/Web-App-Guide.md`

Notes:
- Go and Rust builds are feature-equivalent; pick either.
- Use the `*.sha256` file to verify downloads if desired.

## GHCR (Docker Image) Deployment

The web server is built and pushed to GHCR on every `v*` tag.

Image name:

`ghcr.io/icySparkle/usxtocsv-web:<tag>`

Examples:

```bash
docker pull ghcr.io/icySparkle/usxtocsv-web:v1.0.20260124.4
docker run -p 8080:8080 ghcr.io/icySparkle/usxtocsv-web:v1.0.20260124.4
```

Use `latest` for the most recent tag:

```bash
docker pull ghcr.io/icySparkle/usxtocsv-web:latest
docker run -p 8080:8080 ghcr.io/icySparkle/usxtocsv-web:latest
```

Permissions:
- Public repo => image is public by default.
- If you ever make the repo private, GHCR requires auth to pull.

### Download and Run (Windows)

1) Download the `usxtocsv-go-windows-latest.zip` or `usxtocsv-rust-windows-latest.zip` asset.
2) Extract the zip file.
3) Run from PowerShell:

```powershell
.\usxtocsv.exe -input "C:\Bible\JHN.usx"
```

Verify checksum (optional):

```powershell
Get-FileHash .\usxtocsv-*-windows-latest.zip -Algorithm SHA256
Get-Content .\usxtocsv-*-windows-latest.zip.sha256
```

### Download and Run (macOS)

1) Download the `usxtocsv-go-macos-latest.tar.gz` or `usxtocsv-rust-macos-latest.tar.gz` asset.
2) Extract:

```bash
tar -xzf usxtocsv-*-macos-latest.tar.gz
```

3) Run:

```bash
./usxtocsv -input "/Users/you/Bible/JHN.usx"
```

Verify checksum (optional):

```bash
shasum -a 256 usxtocsv-*-macos-latest.tar.gz
cat usxtocsv-*-macos-latest.tar.gz.sha256
```

### Download and Run (Linux)

1) Download the `usxtocsv-go-ubuntu-latest.tar.gz` or `usxtocsv-rust-ubuntu-latest.tar.gz` asset.
2) Extract:

```bash
tar -xzf usxtocsv-*-ubuntu-latest.tar.gz
```

3) Run:

```bash
./usxtocsv -input "/home/you/Bible/JHN.usx"
```

Verify checksum (optional):

```bash
sha256sum usxtocsv-*-ubuntu-latest.tar.gz
cat usxtocsv-*-ubuntu-latest.tar.gz.sha256
```

---

## 📁 Example CSV Row

```csv
Book,Chapter,Verse,TextPlain,TextStyled,Footnotes,Crossrefs,Subtitle
3JN,1,1,"The elder to the beloved Gaius, whom I love in truth.","The elder to the beloved Gaius, whom I love in truth.",,"","Greeting"
```

If styling is present:

```csv
3JN,1,1,"The elder to the beloved Gaius...","<bdit>The elder</bdit> to the beloved Gaius...",,"","Greeting"
```

---

## 🧠 What the Script Handles

### USX

- `<verse sid="">` / `<verse eid="">` milestones  
- `<para style="">` including headings and body paragraphs  
- `<char style="">` inline styles mapped to tags  
- `<note style="f">`, `<note style="x">` with FT extraction  

### USFM/SFM

- `\id`, `\c`, `\v`  
- Paragraphs: `\m`, `\p`, `\pi`, `\q`, `\q1`–`\q4`, `\qt`, `\qt1`–`\qt4`  
- Headings: `\s`, `\s1`–`\s3`, `\sp`, `\ms`, `\mr`, `\mt`, `\mt1`, `\mt2`  
- Notes: `\f ... \f*`, `\x ... \x*` with FT-only extraction  
- Inline styling: `\bd`, `\it`, `\add`, `\nd`, `\wj`, and their `\+` forms  
- Superscripts: `\sup ... \sup*`, `\+sup ... \+sup*` removed  

---

## 🛠 Requirements

- PowerShell script: Windows PowerShell 5.1 **or** PowerShell 7+  
- Go CLI: Go 1.22+ (to build from source)  
- Rust CLI: Rust 1.70+ and Cargo (to build from source)  
- Web app: Go 1.22+ (local run) or Docker (container run)  
- UTF-8 encoded USX/USFM/SFM files  

---

## 📝 Limitations

- Does not yet export table structures (e.g., `\tr`, `\tc1`, etc.) as structured CSV  
- No reverse-mapping from CSV back to USX/USFM  
- Multi-book USX files are treated as single-book input; multiple `<book>` elements are not merged across files  
- Poetry/paragraph structure (e.g., `q1`, `q2`) is not currently represented as a separate CSV column  

---

## 🔮 Planned Enhancements

- Poetry-level export (`q`, `q1`, `q2`, etc.) into a `PoetryLevel` or `ParaStyle` column  
- Optional companion `Footnotes.csv` and `Crossrefs.csv` with detailed metadata  
- Subtitle index CSV (Book, Chapter, FirstVerse, Subtitle)  
- JSON-based configuration for style mappings and publisher-specific options  

---

## 🤝 Contributing

Contributions are welcome! Helpful contributions include:

- Edge-case USX/USFM test files  
- Performance improvements  
- Publisher-specific style mapping examples  
- Additional export formats (JSON, Parquet, etc.)

---

## 📜 License

Released under the **MIT License** — free for commercial and non-commercial use.

---

## 🙌 Acknowledgements

This tool is inspired by real-world Scripture publishing and translation workflows and is designed to integrate smoothly with:

- Paratext USFM/USX output  
- Digital Bible Library (DBL) USX exports  
- Professional typesetting systems such as Adobe InDesign  

The goal is to make high-quality, verse-structured Scripture data easy to inspect, transform, and typeset.
