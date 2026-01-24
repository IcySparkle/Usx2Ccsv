# CSV Schema

Each row represents a single verse.

## Columns
- **Book**: USX `<book code="">` or USFM `\id` value
- **Chapter**: numeric chapter number
- **Verse**: verse number (supports `1`, `1a`, `1b`, etc.)
- **TextPlain**: verse text with inline styling removed
- **TextStyled**: verse text with inline tags preserved
- **Footnotes**: FT-only footnotes joined with ` | `
- **Crossrefs**: FT-only cross-references joined with ` | `
- **Subtitle**: last seen heading text

## Inline style mapping
- `wj`   -> `<wj>...</wj>`
- `add`  -> `<add>...</add>`
- `nd`   -> `<nd>...</nd>`
- `it`   -> `<i>...</i>`
- `bd`   -> `<b>...</b>`
- `bdit` -> `<bdit>...</bdit>`
- other styles -> `<span>...</span>`

## Notes and behavior
- One CSV row per verse.
- Verse text is merged across paragraph lines.
- Superscripts are removed from both `TextPlain` and `TextStyled`.
- Footnotes and crossrefs include only FT text; markers and callers are ignored.
- Subtitle persists until replaced by a new heading.

## Example row
```csv
Book,Chapter,Verse,TextPlain,TextStyled,Footnotes,Crossrefs,Subtitle
3JN,1,1,"The elder to the beloved Gaius...","<bdit>The elder</bdit> to the beloved Gaius...",,"","Greeting"
```
