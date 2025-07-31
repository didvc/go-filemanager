# File Manager Documentation

## Overview
This is a Go-based web file manager with preview capabilities.

## Features
- 📁 Directory browsing
- 📤 File upload
- 📥 File download
- 🗑️ File/folder deletion
- 👁️ File preview
- 🖼️ Image viewing
- 📄 Text file viewing
- 🌐 HTML rendering

## Supported File Types
### Images
- JPG, JPEG, PNG, GIF, BMP, SVG, WebP

### Text Files
- TXT, MD, JSON, XML, CSV, LOG
- YML, YAML, INI, CONF, CFG

### Code Files
- JS, CSS, GO, PY, JAVA, CPP, C, H
- PHP, RB, RS, TS

### Web Files
- HTML, HTM

## Usage
1. Navigate to http://localhost:8082
2. Browse files and folders
3. Click "Preview" button to view supported files
4. Upload new files using the upload button
5. Create new folders as needed

## Security
- Path traversal protection
- Access limited to designated root directory
- File type validation for previews
