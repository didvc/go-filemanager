# Go Web File Manager

A simple, web-based file manager application built with Go that allows you to manage files and folders through a web interface.

## Features

- Browse directories - Navigate through folders with breadcrumb navigation
- Upload files - Upload files to any directory
- Download files - Download files directly from the web interface
- Create folders - Create new directories
- Delete files/folders - Remove files and directories
- Security - Path traversal protection to keep files within the designated directory
- Responsive design - Works on desktop and mobile devices

## Installation & Usage

1. Prerequisites: Make sure you have Go installed (Go 1.21 or later)

2. Clone or download the project files

3. Run the application:
   ```bash
   go run main.go
   ```

4. Access the file manager:
   Open your web browser and go to `http://localhost:8080`

## File Structure

```
filemanager/
├── main.go          # Main application file
├── go.mod           # Go module file
├── README.md        # This file
└── files/           # Root directory for file management (created automatically)
```

## How to Use

### Navigation
- Click on folder names to enter directories
- Use the "Up" button to go to the parent directory
- Use breadcrumb navigation to jump to any parent directory

### File Operations
- Upload File: Click the "Upload File" button, select a file, and click "Upload"
- Download File: Click the "Download" button next to any file
- Delete: Click the "Delete" button next to any file or folder (confirmation required)

### Folder Operations
- Create Folder: Click the "New Folder" button, enter a name, and click "Create"
- Navigate: Click on any folder name to enter it

## Configuration

### Change Root Directory
By default, files are managed in the `./files` directory. To change this, modify the `rootDir` variable in `main.go`:

```go
rootDir = "/path/to/your/directory"
```

### Change Port
To run on a different port, modify the port in the `ListenAndServe` call:

```go
log.Fatal(http.ListenAndServe(":3000", nil))  // Changes port to 3000
```

## Security Features

- Path Traversal Protection: Prevents access to files outside the designated root directory
- Safe File Operations: All file operations are validated to ensure they stay within bounds
- Input Validation: Form inputs are validated before processing

## API Endpoints

- `GET /` - Main file browser interface
- `POST /upload` - Upload files
- `GET /download` - Download files
- `POST /mkdir` - Create directories
- `POST /delete` - Delete files/directories

## Browser Compatibility

This application works with all modern web browsers including:
- Chrome/Chromium
- Firefox
- Safari
- Edge

## Development

The application is built with:
- Backend: Go (standard library only)
- Frontend: HTML5, CSS3, JavaScript (vanilla)
- Styling: Custom CSS with responsive design

## Troubleshooting

### Common Issues

1. Permission Denied: Make sure the application has read/write permissions to the files directory
2. Port Already in Use: Change the port number in `main.go` if port 8080 is already occupied
3. File Upload Issues: Check that the files directory exists and is writable

### Error Messages

- "Access denied" - Attempting to access files outside the root directory
- "Failed to read directory" - Directory doesn't exist or no read permissions
- "Failed to create file" - No write permissions or disk space issues

## License

This project is open source and available under the MIT License.

<!-- BEGIN gh-mutual-linking -->

---

### Related projects

- [caddy-midi](https://github.com/didvc/caddy-midi): Caddy HTTP handler that serves MIDI files as synthesized audio. Pure Go, no cgo.
- [screen-masking](https://github.com/didvc/screen-masking): Cover parts of your Windows desktop with non-interactive overlays you shape from a pixel-ruled preview window. Pure Win32, no dependencies.
- [simple-ots](https://github.com/didvc/simple-ots): Hash files, build a Merkle tree, anchor to Bitcoin via OpenTimestamps. Selective disclosure without ZKP.
- [rtx-manual-to-md](https://github.com/didvc/rtx-manual-to-md): Convert the Yamaha RTX router command reference HTML archive to GitHub Flavored Markdown, for LLM ingestion, RAG pipelines, and offline browsing.
- [dir-cpu](https://github.com/didvc/dir-cpu): Real-time CLI that shows CPU usage aggregated by filesystem directory
- [image-gallery-app](https://github.com/didvc/image-gallery-app): Modern minimalist image gallery built with Express.js and Vue.js - featuring drag & drop upload, responsive design, and clean aesthetics
- [vibe-go-image-gallery](https://github.com/didvc/vibe-go-image-gallery): Modern image gallery application built with Go and Vue.js featuring SEO optimization, responsive design, and automatic thumbnail generation.
- [go-chatapp-ai](https://github.com/didvc/go-chatapp-ai): A simple, real-time chat application built with Go and WebSockets, featuring a clean web interface for instant messaging. Perfect for learning…
<!-- END gh-mutual-linking -->