# Automation & Integration

This page shows how to use the CLI in automated pipelines and tools like n8n.

## JSON summary output
Use `-json` for machine-readable results:

```bash
./usxtocsv -input "/path/to/FILE.usx" -json
```

Output includes:
- input path
- output path
- format
- rows count

Example output:
```json
{
  "files": [
    {
      "input": "/path/to/FILE.usx",
      "output": "/path/to/FILE.csv",
      "format": "usx",
      "rows": 256
    }
  ]
}
```

## Quiet mode
Suppress progress logs (stdout stays clean for JSON parsing):
```bash
./usxtocsv -input "/path/to/FILE.usx" -quiet -json
```

## Exit codes
- 0: success
- 1: error (invalid input, parse errors, or file issues)

## n8n example
Use an Execute Command node:
```bash
./usxtocsv -input "/data/FILE.usx" -json
```

Parse stdout JSON and use it in downstream nodes.

## Batch processing
```bash
./usxtocsv -input "/data/*.usx" -output "/data/csv" -json
```

## Notes
- Progress logs go to stderr, so stdout stays machine-readable.
- The web app is best for manual use; the CLI is best for automation.
