# Docker and GHCR

The web server image is built and pushed to GHCR on every `v*` tag.

## Image name
`ghcr.io/icysparkle/usxtocsv-web:<tag>`

## Pull and run
```bash
docker pull ghcr.io/icysparkle/usxtocsv-web:v1.0.20260124.15
docker run -p 8080:8080 ghcr.io/icysparkle/usxtocsv-web:v1.0.20260124.15
```

Use `latest` to get the most recent tagged release:
```bash
docker pull ghcr.io/icysparkle/usxtocsv-web:latest
docker run -p 8080:8080 ghcr.io/icysparkle/usxtocsv-web:latest
```

## Ports and endpoints
- Port: `8080`
- React UI: `http://localhost:8080`
- Simple UI: `http://localhost:8080/simple`
- API: `POST /convert`

## Build and run locally
```bash
docker build -t usxtocsv-web .
docker run -p 8080:8080 usxtocsv-web
```

## Common scenarios

Use a different host port:
```bash
docker run -p 9000:8080 ghcr.io/icysparkle/usxtocsv-web:latest
```

Run in detached mode:
```bash
docker run -d -p 8080:8080 ghcr.io/icysparkle/usxtocsv-web:latest
```

## Notes
- Public repo => image is public by default.
- If the repo becomes private, you must `docker login ghcr.io` to pull.
