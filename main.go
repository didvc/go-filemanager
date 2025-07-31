package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type FileInfo struct {
	Name     string
	Path     string
	IsDir    bool
	Size     int64
	ModTime  time.Time
	SizeStr  string
	FileType string
	CanPreview bool
}

type PageData struct {
	CurrentPath string
	ParentPath  string
	Files       []FileInfo
	Breadcrumbs []Breadcrumb
}

type Breadcrumb struct {
	Name string
	Path string
}

var rootDir string

func main() {
	// Set the root directory for file management
	rootDir = "./files"
	
	// Create the files directory if it doesn't exist
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		log.Fatal("Failed to create files directory:", err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/mkdir", mkdirHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/preview", previewHandler)
	http.HandleFunc("/view", viewHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("File Manager Server starting on http://0.0.0.0:8086")
	log.Fatal(http.ListenAndServe("0.0.0.0:8086", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Set UTF-8 encoding
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	// Ensure path is safe and within root directory
	safePath := filepath.Join(rootDir, path)
	safePath = filepath.Clean(safePath)
	
	// Get absolute paths for comparison
	absRootDir, _ := filepath.Abs(rootDir)
	absSafePath, _ := filepath.Abs(safePath)
	
	if !strings.HasPrefix(absSafePath, absRootDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	files, err := readDir(safePath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	// Generate breadcrumbs
	breadcrumbs := generateBreadcrumbs(path)

	// Parent path for navigation
	parentPath := filepath.Dir(path)
	if parentPath == "." {
		parentPath = "/"
	}

	data := PageData{
		CurrentPath: path,
		ParentPath:  parentPath,
		Files:       files,
		Breadcrumbs: breadcrumbs,
	}

	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File Manager</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            border-bottom: 1px solid #ddd;
            padding-bottom: 20px;
            margin-bottom: 20px;
        }
        .breadcrumbs {
            margin: 10px 0;
        }
        .breadcrumbs a {
            color: #007bff;
            text-decoration: none;
            margin-right: 5px;
        }
        .breadcrumbs a:hover {
            text-decoration: underline;
        }
        .actions {
            margin: 20px 0;
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
        }
        .btn-primary {
            background-color: #007bff;
            color: white;
        }
        .btn-success {
            background-color: #28a745;
            color: white;
        }
        .btn-danger {
            background-color: #dc3545;
            color: white;
        }
        .btn:hover {
            opacity: 0.8;
        }
        .file-list {
            width: 100%;
            border-collapse: collapse;
        }
        .file-list th,
        .file-list td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        .file-list th {
            background-color: #f8f9fa;
            font-weight: bold;
        }
        .file-list tr:hover {
            background-color: #f5f5f5;
        }
        .file-icon {
            margin-right: 8px;
        }
        .folder {
            color: #ffc107;
        }
        .file {
            color: #6c757d;
        }
        .file-name {
            color: #007bff;
            text-decoration: none;
        }
        .file-name:hover {
            text-decoration: underline;
        }
        .upload-form {
            margin: 20px 0;
            padding: 20px;
            background-color: #f8f9fa;
            border-radius: 4px;
        }
        .form-group {
            margin-bottom: 15px;
        }
        .form-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
        }
        .form-group input {
            width: 100%;
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
        }
        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0,0,0,0.5);
        }
        .modal-content {
            background-color: white;
            margin: 15% auto;
            padding: 20px;
            border-radius: 8px;
            width: 400px;
            max-width: 90%;
        }
        .close {
            color: #aaa;
            float: right;
            font-size: 28px;
            font-weight: bold;
            cursor: pointer;
        }
        .close:hover {
            color: black;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📁 File Manager</h1>
            <div class="breadcrumbs">
                <strong>Path:</strong>
                {{range .Breadcrumbs}}
                    <a href="/?path={{.Path}}">{{.Name}}</a> /
                {{end}}
            </div>
        </div>

        <div class="actions">
            {{if ne .CurrentPath "/"}}
                <a href="/?path={{.ParentPath}}" class="btn btn-primary">⬆️ Up</a>
            {{end}}
            <button onclick="showUploadModal()" class="btn btn-success">📤 Upload File</button>
            <button onclick="showMkdirModal()" class="btn btn-success">📁 New Folder</button>
        </div>

        <table class="file-list">
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Size</th>
                    <th>Modified</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                {{range .Files}}
                <tr>
                    <td>
                        {{if .IsDir}}
                            <span class="file-icon folder">📁</span>
                            <a href="/?path={{.Path}}" class="file-name">{{.Name}}</a>
                        {{else}}
                            <span class="file-icon file">{{getFileIcon .FileType .IsDir}}</span>
                            <span>{{.Name}}</span>
                        {{end}}
                    </td>
                    <td>{{if not .IsDir}}{{.SizeStr}}{{end}}</td>
                    <td>{{.ModTime.Format "2006-01-02 15:04:05"}}</td>
                    <td>
                        {{if not .IsDir}}
                            {{if .CanPreview}}
                                <a href="/preview?path={{.Path}}" class="btn btn-success" style="padding: 4px 8px; font-size: 12px;" target="_blank">👁️ Preview</a>
                            {{end}}
                            <a href="/download?path={{.Path}}" class="btn btn-primary" style="padding: 4px 8px; font-size: 12px;">📥 Download</a>
                        {{end}}
                        <button onclick="deleteItem('{{.Path}}', '{{.Name}}')" class="btn btn-danger" style="padding: 4px 8px; font-size: 12px;">🗑️ Delete</button>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <!-- Upload Modal -->
    <div id="uploadModal" class="modal">
        <div class="modal-content">
            <span class="close" onclick="hideUploadModal()">&times;</span>
            <h3>Upload File</h3>
            <form action="/upload" method="post" enctype="multipart/form-data">
                <div class="form-group">
                    <label for="file">Choose file:</label>
                    <input type="file" id="file" name="file" required>
                </div>
                <input type="hidden" name="path" value="{{.CurrentPath}}">
                <button type="submit" class="btn btn-success">Upload</button>
            </form>
        </div>
    </div>

    <!-- Create Directory Modal -->
    <div id="mkdirModal" class="modal">
        <div class="modal-content">
            <span class="close" onclick="hideMkdirModal()">&times;</span>
            <h3>Create New Folder</h3>
            <form action="/mkdir" method="post">
                <div class="form-group">
                    <label for="dirname">Folder name:</label>
                    <input type="text" id="dirname" name="dirname" required>
                </div>
                <input type="hidden" name="path" value="{{.CurrentPath}}">
                <button type="submit" class="btn btn-success">Create</button>
            </form>
        </div>
    </div>

    <script>
        function showUploadModal() {
            document.getElementById('uploadModal').style.display = 'block';
        }

        function hideUploadModal() {
            document.getElementById('uploadModal').style.display = 'none';
        }

        function showMkdirModal() {
            document.getElementById('mkdirModal').style.display = 'block';
        }

        function hideMkdirModal() {
            document.getElementById('mkdirModal').style.display = 'none';
        }

        function deleteItem(path, name) {
            if (confirm('Are you sure you want to delete "' + name + '"?')) {
                const form = document.createElement('form');
                form.method = 'POST';
                form.action = '/delete';
                
                const pathInput = document.createElement('input');
                pathInput.type = 'hidden';
                pathInput.name = 'path';
                pathInput.value = path;
                
                form.appendChild(pathInput);
                document.body.appendChild(form);
                form.submit();
            }
        }

        // Close modals when clicking outside
        window.onclick = function(event) {
            const uploadModal = document.getElementById('uploadModal');
            const mkdirModal = document.getElementById('mkdirModal');
            if (event.target == uploadModal) {
                hideUploadModal();
            }
            if (event.target == mkdirModal) {
                hideMkdirModal();
            }
        }
    </script>
</body>
</html>
	`

	t, err := template.New("index").Funcs(template.FuncMap{
		"getFileIcon": getFileIcon,
	}).Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func readDir(path string) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relativePath := strings.TrimPrefix(path, rootDir)
		if relativePath == "" {
			relativePath = "/"
		}
		if !strings.HasSuffix(relativePath, "/") && relativePath != "/" {
			relativePath += "/"
		}
		
		itemPath := relativePath + entry.Name()
		if strings.HasPrefix(itemPath, "//") {
			itemPath = strings.TrimPrefix(itemPath, "/")
		}
		// Remove 'files/' prefix if it exists
		if strings.HasPrefix(itemPath, "files/") {
			itemPath = strings.TrimPrefix(itemPath, "files")
		}

		fileType, canPreview := getFileType(entry.Name())

		file := FileInfo{
			Name:       entry.Name(),
			Path:       itemPath,
			IsDir:      entry.IsDir(),
			Size:       info.Size(),
			ModTime:    info.ModTime(),
			SizeStr:    formatSize(info.Size()),
			FileType:   fileType,
			CanPreview: canPreview,
		}
		files = append(files, file)
	}

	// Sort: directories first, then files, both alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return files, nil
}

func generateBreadcrumbs(path string) []Breadcrumb {
	var breadcrumbs []Breadcrumb
	
	breadcrumbs = append(breadcrumbs, Breadcrumb{Name: "Home", Path: "/"})
	
	if path != "/" {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		currentPath := ""
		for _, part := range parts {
			if part != "" {
				currentPath += "/" + part
				breadcrumbs = append(breadcrumbs, Breadcrumb{Name: part, Path: currentPath})
			}
		}
	}
	
	return breadcrumbs
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func getFileType(filename string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt", ".md", ".json", ".xml", ".csv", ".log", ".yml", ".yaml", ".ini", ".conf", ".cfg":
		return "text", true
	case ".html", ".htm":
		return "html", true
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp":
		return "image", true
	case ".pdf":
		return "pdf", true
	case ".js", ".css", ".go", ".py", ".java", ".cpp", ".c", ".h", ".php", ".rb", ".rs", ".ts":
		return "code", true
	default:
		return "binary", false
	}
}

func getFileIcon(fileType string, isDir bool) string {
	if isDir {
		return "📁"
	}
	switch fileType {
	case "text":
		return "📄"
	case "html":
		return "🌐"
	case "image":
		return "🖼️"
	case "pdf":
		return "📕"
	case "code":
		return "💻"
	default:
		return "📄"
	}
}

func isValidUTF8(data []byte) bool {
	return utf8.Valid(data)
}

func sanitizeUTF8(data []byte) string {
	// If valid UTF-8, return as string
	if utf8.Valid(data) {
		return string(data)
	}
	
	// If not valid UTF-8, convert invalid sequences to replacement character
	return strings.ToValidUTF8(string(data), "�")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.FormValue("path")
	if path == "" {
		path = "/"
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create the destination path
	destDir := filepath.Join(rootDir, path)
	destPath := filepath.Join(destDir, header.Filename)

	// Ensure the destination is within the root directory
	absRootDir, _ := filepath.Abs(rootDir)
	absDestPath, _ := filepath.Abs(filepath.Clean(destPath))
	
	if !strings.HasPrefix(absDestPath, absRootDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Create the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Copy the uploaded file to the destination
	_, err = io.Copy(destFile, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Redirect back to the directory
	http.Redirect(w, r, "/?path="+path, http.StatusSeeOther)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path not specified", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(rootDir, path)
	filePath = filepath.Clean(filePath)

	// Ensure the file is within the root directory
	absRootDir, _ := filepath.Abs(rootDir)
	absFilePath, _ := filepath.Abs(filePath)
	
	if !strings.HasPrefix(absFilePath, absRootDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check if file exists and is not a directory
	info, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		http.Error(w, "Cannot download directory", http.StatusBadRequest)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(w, r, filePath)
}

func mkdirHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.FormValue("path")
	dirname := r.FormValue("dirname")

	if dirname == "" {
		http.Error(w, "Directory name not specified", http.StatusBadRequest)
		return
	}

	if path == "" {
		path = "/"
	}

	// Create the new directory path
	newDirPath := filepath.Join(rootDir, path, dirname)
	newDirPath = filepath.Clean(newDirPath)

	// Ensure the directory is within the root directory
	absRootDir, _ := filepath.Abs(rootDir)
	absNewDirPath, _ := filepath.Abs(newDirPath)
	
	if !strings.HasPrefix(absNewDirPath, absRootDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Create the directory
	err := os.MkdirAll(newDirPath, 0755)
	if err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Redirect back to the directory
	http.Redirect(w, r, "/?path="+path, http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.FormValue("path")
	if path == "" {
		http.Error(w, "Path not specified", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(rootDir, path)
	filePath = filepath.Clean(filePath)

	// Ensure the file/directory is within the root directory
	absRootDir, _ := filepath.Abs(rootDir)
	absFilePath, _ := filepath.Abs(filePath)
	
	if !strings.HasPrefix(absFilePath, absRootDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete the file or directory
	err := os.RemoveAll(filePath)
	if err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	// Get the parent directory for redirect
	parentPath := filepath.Dir(path)
	if parentPath == "." {
		parentPath = "/"
	}

	// Redirect back to the parent directory
	http.Redirect(w, r, "/?path="+parentPath, http.StatusSeeOther)
}

func previewHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path not specified", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(rootDir, path)
	filePath = filepath.Clean(filePath)

	// Ensure the file is within the root directory
	absRootDir, _ := filepath.Abs(rootDir)
	absFilePath, _ := filepath.Abs(filePath)
	
	if !strings.HasPrefix(absFilePath, absRootDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check if file exists and is not a directory
	info, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		http.Error(w, "Cannot preview directory", http.StatusBadRequest)
		return
	}

	filename := filepath.Base(filePath)
	fileType, canPreview := getFileType(filename)

	if !canPreview {
		http.Error(w, "File type not supported for preview", http.StatusUnsupportedMediaType)
		return
	}

	switch fileType {
	case "image":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		imagePreviewTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Image Preview - %s</title>
    <meta charset="UTF-8">
    <style>
        body { margin: 0; padding: 20px; background: #f0f0f0; font-family: Arial, sans-serif; }
        .container { max-width: 100%%; text-align: center; }
        img { max-width: 100%%; max-height: 90vh; box-shadow: 0 4px 8px rgba(0,0,0,0.1); }
        .info { margin: 20px 0; }
        .btn { padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="info">
            <h2>🖼️ %s</h2>
            <p>Size: %s</p>
        </div>
        <img src="/view?path=%s" alt="%s">
        <div style="margin-top: 20px;">
            <a href="/download?path=%s" class="btn">📥 Download</a>
            <a href="javascript:history.back()" class="btn">← Back</a>
        </div>
    </div>
</body>
</html>`
		fmt.Fprintf(w, imagePreviewTemplate, filename, filename, formatSize(info.Size()), path, filename, path)

	case "text", "code":
		content, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		// Validate UTF-8 and handle encoding
		if !isValidUTF8(content) {
			// If not valid UTF-8, show as binary
			http.Error(w, "File contains non-UTF-8 content. Cannot preview as text.", http.StatusUnsupportedMediaType)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		textPreviewTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Text Preview - %s</title>
    <meta charset="UTF-8">
    <style>
        body { margin: 0; padding: 20px; background: #f8f9fa; font-family: Arial, sans-serif; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { border-bottom: 1px solid #ddd; padding-bottom: 10px; margin-bottom: 20px; }
        .content { background: #f8f9fa; padding: 15px; border-radius: 4px; border: 1px solid #e9ecef; }
        pre { margin: 0; white-space: pre-wrap; word-wrap: break-word; font-family: 'Courier New', monospace; }
        .btn { padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin: 5px; }
        .info { color: #6c757d; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>📄 %s</h2>
            <div class="info">Size: %s | Type: %s | Encoding: UTF-8</div>
        </div>
        <div class="content">
            <pre>%s</pre>
        </div>
        <div style="margin-top: 20px;">
            <a href="/download?path=%s" class="btn">📥 Download</a>
            <a href="javascript:history.back()" class="btn">← Back</a>
        </div>
    </div>
</body>
</html>`
		// Sanitize content to ensure valid UTF-8 and escape HTML
		sanitizedContent := sanitizeUTF8(content)
		escapedContent := strings.ReplaceAll(sanitizedContent, "<", "&lt;")
		escapedContent = strings.ReplaceAll(escapedContent, ">", "&gt;")
		fmt.Fprintf(w, textPreviewTemplate, filename, filename, formatSize(info.Size()), fileType, escapedContent, path)

	case "html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		htmlPreviewTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>HTML Preview - %s</title>
    <meta charset="UTF-8">
    <style>
        body { margin: 0; padding: 20px; background: #f8f9fa; font-family: Arial, sans-serif; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: white; padding: 15px; margin-bottom: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .preview-frame { width: 100%%; height: 70vh; border: 1px solid #ddd; border-radius: 4px; }
        .btn { padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin: 5px; }
        .info { color: #6c757d; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>🌐 %s</h2>
            <div class="info">Size: %s | Encoding: UTF-8</div>
            <div style="margin-top: 10px;">
                <a href="/view?path=%s" class="btn" target="_blank">🔗 Open in New Tab</a>
                <a href="/download?path=%s" class="btn">📥 Download</a>
                <a href="javascript:history.back()" class="btn">← Back</a>
            </div>
        </div>
        <iframe src="/view?path=%s" class="preview-frame"></iframe>
    </div>
</body>
</html>`
		fmt.Fprintf(w, htmlPreviewTemplate, filename, filename, formatSize(info.Size()), path, path, path)

	default:
		http.Error(w, "File type not supported for preview", http.StatusUnsupportedMediaType)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path not specified", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(rootDir, path)
	filePath = filepath.Clean(filePath)

	// Ensure the file is within the root directory
	absRootDir, _ := filepath.Abs(rootDir)
	absFilePath, _ := filepath.Abs(filePath)
	
	if !strings.HasPrefix(absFilePath, absRootDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Serve the file directly
	http.ServeFile(w, r, filePath)
}
