# Web App Guide

The web server accepts uploads, converts them to CSV, and returns a zip file.

## How it works
1) You upload one or more files (or a zip).
2) The server converts supported files into CSV.
3) A zip of CSVs is returned to your browser.

## Endpoints
- `/` React UI (default)
- `/simple` built-in minimal HTML
- `/convert` POST upload endpoint

## Accepted uploads
- `.usx`, `.usfm`, `.sfm`
- `.zip` containing one or more of the above

## Limits and behavior
- Max upload size is 200 MB per request.
- Nested folders in zip are flattened by filename.
- Only supported extensions are converted.

## Run locally (Go)
```bash
cd go
go run ./web
```

Open `http://localhost:8080`.

## Run via Docker
```bash
docker run -p 8080:8080 ghcr.io/icysparkle/usxtocsv-web:latest
```

Open:
- React UI: `http://localhost:8080`
- Simple UI: `http://localhost:8080/simple`

## React UI
The React UI is in `web-ui/` and calls `/convert`.

Run the UI locally:
```bash
cd web-ui
npm install
npm run dev
```

Point the UI at a remote backend:
```bash
VITE_API_BASE="https://your-backend.example.com" npm run dev
```

## Calling the API directly

Example using curl:
```bash
curl -X POST \
  -F "files=@/path/to/MAT.usx" \
  -F "files=@/path/to/MRK.usfm" \
  http://localhost:8080/convert \
  --output usxtocsv-output.zip
```

## Common errors
- "No files uploaded": you sent an empty form.
- "No .usx, .usfm, or .sfm files found": upload unsupported files.
- "Failed to parse upload": the request exceeded the size limit.
