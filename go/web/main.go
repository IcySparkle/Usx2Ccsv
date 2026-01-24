package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"usxtocsv/convert"
)

const (
	maxUploadSize = 200 << 20
)

func main() {
	port := flag.String("port", "", "Port to listen on (overrides PORT env)")
	flag.Parse()

	listenPort := resolvePort(*port)

	mux := http.NewServeMux()
	if staticDir := resolveStaticDir(); staticDir != "" {
		mux.Handle("/", spaHandler(staticDir))
		mux.HandleFunc("/simple", handleIndex)
	} else {
		mux.HandleFunc("/", handleIndex)
	}
	mux.HandleFunc("/convert", handleConvert)

	server := &http.Server{
		Addr:              ":" + listenPort,
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Printf("Listening on http://localhost:%s\n", listenPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func resolvePort(flagPort string) string {
	if flagPort != "" {
		return flagPort
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}
	return "8080"
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, indexHTML)
}

func resolveStaticDir() string {
	if dir := os.Getenv("WEB_UI_DIR"); dir != "" {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return ""
}

func spaHandler(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		path := filepath.Clean(r.URL.Path)
		if path == "/" {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}

		full := filepath.Join(dir, path)
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			http.ServeFile(w, r, full)
			return
		}

		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Failed to parse upload", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	tempDir, err := os.MkdirTemp("", "usxtocsv-upload-*")
	if err != nil {
		http.Error(w, "Failed to create temp directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	inputPaths, err := saveUploads(tempDir, files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	inputPaths = filterSupported(inputPaths)
	if len(inputPaths) == 0 {
		http.Error(w, "No .usx, .usfm, or .sfm files found in upload", http.StatusBadRequest)
		return
	}

	outputDir := filepath.Join(tempDir, "out")
	if _, err := convert.ConvertFiles(inputPaths, outputDir, convert.Options{Quiet: true}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=usxtocsv-output.zip")
	w.WriteHeader(http.StatusOK)

	if err := writeZip(w, outputDir); err != nil {
		http.Error(w, "Failed to build zip", http.StatusInternalServerError)
		return
	}
}

func saveUploads(baseDir string, files []*multipart.FileHeader) ([]string, error) {
	var paths []string

	for _, fh := range files {
		if fh == nil {
			continue
		}
		name := sanitizeFilename(fh.Filename)
		if name == "" {
			continue
		}

		src, err := fh.Open()
		if err != nil {
			return nil, fmt.Errorf("Failed to read upload: %s", name)
		}
		defer src.Close()

		destPath := filepath.Join(baseDir, name)
		dest, err := os.Create(destPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to write upload: %s", name)
		}

		if _, err := io.Copy(dest, src); err != nil {
			dest.Close()
			return nil, fmt.Errorf("Failed to save upload: %s", name)
		}
		dest.Close()

		if strings.HasSuffix(strings.ToLower(name), ".zip") {
			extracted, err := extractZip(destPath, baseDir)
			if err != nil {
				return nil, err
			}
			paths = append(paths, extracted...)
			continue
		}

		paths = append(paths, destPath)
	}

	return paths, nil
}

func extractZip(zipPath, destDir string) ([]string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open zip: %s", filepath.Base(zipPath))
	}
	defer reader.Close()

	var extracted []string
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		name := sanitizeFilename(file.Name)
		if name == "" {
			continue
		}

		targetPath := filepath.Join(destDir, name)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return nil, fmt.Errorf("Failed to extract zip: %s", filepath.Base(zipPath))
		}

		src, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("Failed to extract zip: %s", filepath.Base(zipPath))
		}
		defer src.Close()

		dest, err := os.Create(targetPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to extract zip: %s", filepath.Base(zipPath))
		}

		if _, err := io.Copy(dest, src); err != nil {
			dest.Close()
			return nil, fmt.Errorf("Failed to extract zip: %s", filepath.Base(zipPath))
		}
		dest.Close()
		extracted = append(extracted, targetPath)
	}

	return extracted, nil
}

func filterSupported(paths []string) []string {
	var filtered []string
	for _, path := range paths {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".usx", ".usfm", ".sfm":
			filtered = append(filtered, path)
		}
	}
	return filtered
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, string(filepath.Separator), "_")
	name = strings.TrimSpace(name)
	return name
}

func writeZip(w io.Writer, dir string) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if strings.ToLower(filepath.Ext(path)) != ".csv" {
			continue
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		wr, err := zipWriter.Create(entry.Name())
		if err != nil {
			return err
		}
		if _, err := io.Copy(wr, file); err != nil {
			return err
		}
	}

	return nil
}

const indexHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>USX/USFM to CSV</title>
    <style>
      :root {
        --bg: #f4f1ec;
        --ink: #1f1b16;
        --accent: #d46b08;
        --panel: #fffaf3;
        --border: #e7dccb;
      }
      body {
        margin: 0;
        font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
        background: radial-gradient(circle at 10% 10%, #fff1d6, #f4f1ec);
        color: var(--ink);
      }
      .wrap {
        max-width: 720px;
        margin: 48px auto;
        background: var(--panel);
        border: 1px solid var(--border);
        border-radius: 16px;
        padding: 32px;
        box-shadow: 0 16px 40px rgba(36, 25, 12, 0.12);
      }
      h1 {
        margin: 0 0 12px;
        font-size: 28px;
        letter-spacing: 0.5px;
      }
      p {
        margin: 0 0 20px;
        line-height: 1.6;
      }
      .drop {
        border: 2px dashed var(--border);
        border-radius: 12px;
        padding: 24px;
        background: #fffdf9;
      }
      .btn {
        display: inline-block;
        margin-top: 16px;
        padding: 10px 18px;
        background: var(--accent);
        color: #fff;
        border: none;
        border-radius: 8px;
        cursor: pointer;
      }
      .note {
        margin-top: 12px;
        font-size: 13px;
        color: #5f5245;
      }
    </style>
  </head>
  <body>
    <div class="wrap">
      <h1>USX / USFM / SFM to CSV</h1>
      <p>Upload one or more files, or a zip containing multiple files. The server returns a zip of CSVs.</p>
      <form class="drop" action="/convert" method="post" enctype="multipart/form-data">
        <input type="file" name="files" multiple />
        <div class="note">Accepted: .usx, .usfm, .sfm, or .zip</div>
        <button class="btn" type="submit">Convert</button>
      </form>
    </div>
  </body>
</html>`
