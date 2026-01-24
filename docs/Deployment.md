# Deployment

This guide explains how to deploy the web server with Docker.

## Render
1) Create a new Web Service from your GitHub repo.
2) Environment: Docker.
3) Set port to `8080`.
4) Deploy.

Optional:
- Add an environment variable `PORT=8080` (Render usually sets this automatically).

## Fly.io

Initialize and deploy:
```bash
fly launch
fly deploy
```

Make sure `PORT=8080` is set in your Fly app configuration.

## Railway
1) New Project -> Deploy from GitHub.
2) Select Dockerfile.
3) Set port to `8080` if prompted.
4) Deploy.

## DigitalOcean App Platform
1) Create App from GitHub.
2) Use Dockerfile.
3) Set HTTP port to `8080`.
4) Deploy.

## After deployment
- Open the service URL and confirm the React UI loads.
- Visit `/simple` to confirm the fallback HTML works.
- Test `/convert` with a small file.
