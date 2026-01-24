import { useMemo, useState } from "react";

const API_BASE = import.meta.env.VITE_API_BASE || "";

export default function App() {
  const [files, setFiles] = useState([]);
  const [status, setStatus] = useState("idle");
  const [error, setError] = useState("");
  const [downloadUrl, setDownloadUrl] = useState("");

  const acceptedList = useMemo(
    () => ".usx,.usfm,.sfm,.zip",
    []
  );

  const handleFilesChange = (event) => {
    setError("");
    setDownloadUrl("");
    setFiles(Array.from(event.target.files || []));
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    setDownloadUrl("");

    if (!files.length) {
      setError("Please choose at least one file.");
      return;
    }

    const formData = new FormData();
    files.forEach((file) => formData.append("files", file));

    try {
      setStatus("uploading");
      const response = await fetch(`${API_BASE}/convert`, {
        method: "POST",
        body: formData
      });

      if (!response.ok) {
        const text = await response.text();
        throw new Error(text || "Conversion failed.");
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      setDownloadUrl(url);
      setStatus("done");
    } catch (err) {
      setStatus("error");
      setError(err.message || "Conversion failed.");
    }
  };

  return (
    <div className="page">
      <header className="hero">
        <div className="hero__pill">USX / USFM / SFM</div>
        <h1>Turn scripture sources into clean CSVs.</h1>
        <p>
          Upload files or a zip bundle. The converter returns a zip with
          one CSV per input file.
        </p>
      </header>

      <main className="panel">
        <form onSubmit={handleSubmit} className="panel__form">
          <label className="dropzone">
            <input
              type="file"
              multiple
              accept={acceptedList}
              onChange={handleFilesChange}
            />
            <div>
              <strong>Drop files here</strong> or browse
            </div>
            <span>Accepted: .usx .usfm .sfm .zip</span>
          </label>

          <div className="filelist">
            {files.length ? (
              files.map((file) => (
                <div key={file.name} className="filelist__item">
                  {file.name}
                </div>
              ))
            ) : (
              <div className="filelist__empty">No files selected.</div>
            )}
          </div>

          <button className="action" type="submit" disabled={status === "uploading"}>
            {status === "uploading" ? "Converting..." : "Convert to CSV"}
          </button>

          {status === "done" && downloadUrl && (
            <a className="download" href={downloadUrl} download="usxtocsv-output.zip">
              Download CSVs
            </a>
          )}

          {status === "error" && <div className="error">{error}</div>}
          {status === "idle" && (
            <div className="hint">
              Tip: you can upload a zip that contains multiple source files.
            </div>
          )}
        </form>
      </main>

      <footer className="footer">
        <span>Need automation? Use the CLI or API directly.</span>
      </footer>
    </div>
  );
}
