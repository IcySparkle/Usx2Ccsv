# Troubleshooting

Common issues and fixes.

## "Input must be a .usx, .usfm, or .sfm file"
Cause: The input path is empty, unsupported, or points to a file with a different extension.
Fix: Use `.usx`, `.usfm`, `.sfm`, or a folder containing those files.

## "No files found"
Cause: The folder or wildcard did not match any supported files.
Fix: Confirm the path and extension. Try a direct file path first.

## "Input path not found"
Cause: The path is misspelled or the file does not exist.
Fix: Verify the file exists and is accessible from the current working directory.

## Release artifacts missing
Cause: The release workflow failed or was skipped for the tag.
Fix:
- Check GitHub Actions for the tag run.
- Confirm the `release` job completed successfully.

## Web server returns 500
Cause: Server-side conversion failed.
Fix:
- Confirm the upload is a supported file or zip.
- Check the server logs for the exact error message.

## Web UI loads but /convert fails
Cause: Backend URL mismatch or CORS issues.
Fix:
- Ensure the UI points to the correct API base.
- Use the Docker image or same-origin setup during dev.

## Docker build fails
Cause: Missing build context or dependency issues.
Fix:
- Make sure the Docker build context includes `web-ui/` and `go/`.
- Rebuild with `--no-cache` to verify dependency downloads.

## GHCR image pull fails
Cause: Wrong image name or authentication required.
Fix:
- Check the exact tag in GitHub Releases.
- If repo is private, `docker login ghcr.io` before pulling.
