# UsxToCsv — USX → CSV Conversion Tool
### PowerShell script for converting USX format to CSV.

---

## 📌 Overview

`UsxToCsv.ps1` converts **USX (Unified Scripture XML)** files into clean, structured **CSV** files suitable for:

- Bible publishing pipelines  
- InDesign Data Merge  
- Text alignment and search systems  
- Scripture QA and linguistic workflows  
- Multi-version parallel Bibles  

It handles complex USX constructs, including verse milestones, inline character markup, notes, and subtitles.

---

## ✨ Key Features

### ✔ Converts USX to CSV (1 row per verse)  
Supports both **single file** and **entire folder** processing.

### ✔ Clean, normalized text  
- `TextPlain`: no inline markup  
- `TextStyled`: simplified tags (`<wj>`, `<add>`, `<nd>`, `<i>`, `<b>`, etc.)

### ✔ FT-only footnote extraction  
Extracts **only** `<char style="ft">` content.  
Ignores:
- `caller="+"`
- `fr` markers
- note metadata  
- nested USX markup is flattened into readable text.

### ✔ Subtitle (Pericope Heading) Support  
Recognizes heading styles:
```
s, s1, s2, s3, sp, ms, mr, mt, mt1, mt2
```
The latest subtitle applies to all following verses.

### ✔ InDesign-friendly output  
`TextStyled` is optimized for GREP styling.

---

## 📂 CSV Output Columns

| Column | Details |
|--------|---------|
| `Book` | Book code (GEN, EXO, JHN, etc.) |
| `Chapter` | Chapter number |
| `Verse` | Verse number |
| `TextPlain` | Raw readable text |
| `TextStyled` | Inline formatted text |
| `Footnotes` | FT-only text extracted from footnotes |
| `Crossrefs` | FT-only cross-reference entries |
| `Subtitle` | Section heading above the verse |

---

## 🧠 How It Works

### 1. Verse Detection  
Uses USX milestone structure:

```xml
<verse sid="JHN 3:16" number="16"/>
...
<verse eid="JHN 3:16"/>
```

Text between the two is treated as one verse.

---

### 2. Inline Formatting  
USX:

```xml
<char style="wj">Jesus said…</char>
```

Becomes:

```text
<wj>Jesus said…</wj>
```

---

### 3. Footnote Extraction  
USX:

```xml
<note style="f">
  <char style="fr">1:11</char>
  <char style="ft"><ref loc="ISA 50:9">Is 50:9</ref></char>
</note>
```

Output:

```
Is 50:9
```

---

### 4. Subtitle Assignment  

```xml
<para style="s1">The Birth of Jesus</para>
```

Applies to all verses until another heading appears.

---

## 🚀 Usage

### Convert a single USX file
```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\JHN.usx"
```

### Convert all USX in a folder
```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\USX"
```

### Use a custom output directory
```powershell
.\UsxToCsv.ps1 -InputPath "C:\Bible\USX" -OutputFolder "C:\Bible\CSV"
```

---

## 🔧 Script Architecture

Core functions:

- `Get-AttrValue` — Safe XML attribute getter  
- `Find-FtNode` — Locates `<char style="ft">` inside notes  
- `ExtractFTFromNote` — Returns FT text only  
- `Get-PlainInnerText` — Cleans heading text  
- `Get-StyledTagName` — Maps USX inline styles  
- `Process-Node` — Main XML walker  
- `Process-NoteNode` — Handles footnotes & crossrefs  
- `Add-CurrentVerse` — Builds final CSV row  

---

## 🧪 Testing

Recommended books:
- 1 John  
- John  
- Romans  

Check:
- Verse boundaries  
- Subtitle propagation  
- FT-only note extraction  
- Inline tag formatting  

---

## 🧭 Roadmap

Future enhancements (optional):

- Poetry indentation (`q1`, `q2`, `q3`)  
- Subtitle index CSV  
- Parallel Bible merging  
- Configurable inline-style mapping (JSON)  
- USFM export option  

---

## 📄 License

MIT License — free for ministry, research, publishing, and commercial use.

---

## 💬 Contributions & Support

If you need:
- new features  
- workflow integration  
- script customization  

Feel free to ask or open an issue.

