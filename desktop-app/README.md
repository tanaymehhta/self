# Self Desktop App

Tauri-based desktop application for local file monitoring and audio recording.

## Features

- **Audio Recording**: High-quality audio capture with system microphone
- **File Monitoring**: Real-time file system event tracking
- **System Integration**: System tray, global hotkeys, notifications
- **Cross-Platform**: Windows, macOS, and Linux support
- **Security**: Sandboxed execution with minimal permissions
- **Auto-Updates**: Seamless delta updates for new versions

## Tech Stack

- **Framework**: Tauri 2.0
- **Backend**: Rust for system-level operations
- **Frontend**: HTML/CSS/JavaScript (embedded web view)
- **Audio**: cpal for cross-platform audio recording
- **File Watching**: notify-rs for file system events
- **HTTP Client**: reqwest for API communication
- **Serialization**: serde for JSON handling

## Getting Started

### Prerequisites

- **Rust**: 1.70+ with cargo
- **Node.js**: 20+ (for frontend assets)
- **System Dependencies**:
  - macOS: Xcode Command Line Tools
  - Windows: Microsoft C++ Build Tools
  - Linux: `build-essential`, `libwebkit2gtk-4.0-dev`

### Development Setup

```bash
# Install Tauri CLI
cargo install tauri-cli@^2.0.0

# Install frontend dependencies
npm install

# Start development mode
cargo tauri dev

# Build for production
cargo tauri build
```

## Project Structure

```
desktop-app/
├── src/                   # Rust source code
│   ├── main.rs           # Main application entry
│   ├── audio/            # Audio recording module
│   ├── files/            # File monitoring module
│   ├── api/              # API communication
│   └── system/           # System integration
├── src-tauri/            # Tauri configuration
│   ├── Cargo.toml        # Rust dependencies
│   ├── tauri.conf.json   # Tauri settings
│   └── capabilities/     # Permission definitions
├── ui/                   # Frontend assets
│   ├── index.html        # Main UI
│   ├── styles.css        # Styling
│   └── script.js         # JavaScript logic
└── icons/                # Application icons
```

## Core Modules

### Audio Recording

```rust
use cpal::traits::{DeviceTrait, HostTrait, StreamTrait};

pub struct AudioRecorder {
    stream: Option<cpal::Stream>,
    config: cpal::StreamConfig,
}

impl AudioRecorder {
    pub fn new() -> Result<Self, AudioError> {
        let host = cpal::default_host();
        let device = host.default_input_device()
            .ok_or(AudioError::NoInputDevice)?;

        let config = device.default_input_config()?.into();

        Ok(Self {
            stream: None,
            config,
        })
    }

    pub fn start_recording(&mut self) -> Result<(), AudioError> {
        // Implementation for starting audio recording
    }

    pub fn stop_recording(&mut self) -> Result<Vec<u8>, AudioError> {
        // Implementation for stopping and retrieving audio data
    }
}
```

### File Monitoring

```rust
use notify::{RecommendedWatcher, RecursiveMode, Watcher, Event};

pub struct FileMonitor {
    watcher: RecommendedWatcher,
    watched_paths: Vec<PathBuf>,
}

impl FileMonitor {
    pub fn new(callback: impl Fn(FileEvent) + Send + 'static) -> Result<Self, FileError> {
        let watcher = notify::recommended_watcher(move |res: Result<Event, _>| {
            match res {
                Ok(event) => callback(FileEvent::from(event)),
                Err(e) => eprintln!("Watch error: {:?}", e),
            }
        })?;

        Ok(Self {
            watcher,
            watched_paths: Vec::new(),
        })
    }

    pub fn watch_path(&mut self, path: PathBuf) -> Result<(), FileError> {
        self.watcher.watch(&path, RecursiveMode::Recursive)?;
        self.watched_paths.push(path);
        Ok(())
    }
}
```

### API Communication

```rust
use reqwest::Client;
use serde::{Deserialize, Serialize};

pub struct ApiClient {
    client: Client,
    base_url: String,
    auth_token: Option<String>,
}

impl ApiClient {
    pub fn new(base_url: String) -> Self {
        Self {
            client: Client::new(),
            base_url,
            auth_token: None,
        }
    }

    pub async fn upload_audio(&self, audio_data: Vec<u8>) -> Result<UploadResponse, ApiError> {
        let form = reqwest::multipart::Form::new()
            .part("audio", reqwest::multipart::Part::bytes(audio_data)
                .file_name("recording.wav")
                .mime_str("audio/wav")?);

        let response = self.client
            .post(&format!("{}/api/v1/audio/upload", self.base_url))
            .multipart(form)
            .send()
            .await?;

        Ok(response.json().await?)
    }
}
```

## Tauri Commands

### Audio Commands

```rust
#[tauri::command]
async fn start_recording(state: tauri::State<'_, AppState>) -> Result<(), String> {
    let mut recorder = state.audio_recorder.lock().await;
    recorder.start_recording()
        .map_err(|e| format!("Failed to start recording: {}", e))?;
    Ok(())
}

#[tauri::command]
async fn stop_recording(state: tauri::State<'_, AppState>) -> Result<String, String> {
    let mut recorder = state.audio_recorder.lock().await;
    let audio_data = recorder.stop_recording()
        .map_err(|e| format!("Failed to stop recording: {}", e))?;

    // Upload to server
    let api_client = &state.api_client;
    let response = api_client.upload_audio(audio_data).await
        .map_err(|e| format!("Failed to upload audio: {}", e))?;

    Ok(response.transcription_id)
}
```

### File Commands

```rust
#[tauri::command]
async fn add_watched_folder(
    path: String,
    state: tauri::State<'_, AppState>
) -> Result<(), String> {
    let mut monitor = state.file_monitor.lock().await;
    let path_buf = PathBuf::from(path);

    monitor.watch_path(path_buf)
        .map_err(|e| format!("Failed to watch folder: {}", e))?;

    Ok(())
}

#[tauri::command]
async fn get_recent_files(
    limit: usize,
    state: tauri::State<'_, AppState>
) -> Result<Vec<FileInfo>, String> {
    let monitor = state.file_monitor.lock().await;
    Ok(monitor.get_recent_files(limit))
}
```

## Frontend Integration

### JavaScript API

```javascript
// Audio recording
async function startRecording() {
    try {
        await window.__TAURI__.invoke('start_recording');
        updateUIRecording(true);
    } catch (error) {
        console.error('Failed to start recording:', error);
        showErrorNotification('Recording failed to start');
    }
}

async function stopRecording() {
    try {
        const transcriptionId = await window.__TAURI__.invoke('stop_recording');
        updateUIRecording(false);
        showSuccessNotification('Recording uploaded successfully');
        return transcriptionId;
    } catch (error) {
        console.error('Failed to stop recording:', error);
        showErrorNotification('Recording failed to upload');
    }
}

// File monitoring
async function addWatchedFolder() {
    try {
        const folderPath = await window.__TAURI__.dialog.open({
            directory: true,
            multiple: false
        });

        if (folderPath) {
            await window.__TAURI__.invoke('add_watched_folder', { path: folderPath });
            updateWatchedFolders();
        }
    } catch (error) {
        console.error('Failed to add watched folder:', error);
    }
}
```

### Event Listeners

```javascript
// Listen for file events
window.__TAURI__.event.listen('file-changed', (event) => {
    const fileEvent = event.payload;
    console.log(`File ${fileEvent.action}: ${fileEvent.path}`);
    updateFileList();
});

// Listen for transcription updates
window.__TAURI__.event.listen('transcription-update', (event) => {
    const update = event.payload;
    updateTranscriptionDisplay(update.transcription_id, update.text);
});
```

## Configuration

### Tauri Config

```json
{
  "build": {
    "beforeBuildCommand": "npm run build",
    "beforeDevCommand": "npm run dev",
    "devPath": "../ui",
    "distDir": "../ui/dist"
  },
  "package": {
    "productName": "Self",
    "version": "0.1.0"
  },
  "tauri": {
    "allowlist": {
      "all": false,
      "fs": {
        "readFile": true,
        "readDir": true,
        "scope": ["$HOME/**"]
      },
      "dialog": {
        "open": true,
        "save": true
      },
      "notification": {
        "all": true
      },
      "globalShortcut": {
        "all": true
      }
    },
    "bundle": {
      "active": true,
      "identifier": "com.self.app",
      "targets": "all",
      "icon": [
        "icons/32x32.png",
        "icons/128x128.png",
        "icons/icon.icns",
        "icons/icon.ico"
      ]
    },
    "security": {
      "csp": null
    },
    "updater": {
      "active": true,
      "endpoints": ["https://releases.self-app.com/{{target}}/{{current_version}}"],
      "dialog": true,
      "pubkey": "YOUR_PUBLIC_KEY_HERE"
    },
    "windows": [
      {
        "fullscreen": false,
        "height": 600,
        "resizable": true,
        "title": "Self",
        "width": 800,
        "minHeight": 400,
        "minWidth": 600
      }
    ],
    "systemTray": {
      "iconPath": "icons/icon.png",
      "iconAsTemplate": true
    }
  }
}
```

## Permissions

### Capabilities

```json
{
  "identifier": "self-app-capabilities",
  "description": "Permissions for Self desktop app",
  "local": true,
  "windows": ["main"],
  "permissions": [
    "core:default",
    "dialog:default",
    "fs:default",
    "notification:default",
    "global-shortcut:default",
    "os:default"
  ]
}
```

## Building & Distribution

### Development Build

```bash
# Debug build
cargo tauri dev

# Release build
cargo tauri build

# Build for specific target
cargo tauri build --target x86_64-pc-windows-msvc
```

### Code Signing

```bash
# macOS
export APPLE_CERTIFICATE="Developer ID Application: ..."
export APPLE_CERTIFICATE_PASSWORD="..."
cargo tauri build

# Windows
signtool sign /f certificate.p12 /p password target/release/self.exe
```

### Auto Updates

```rust
// Check for updates
use tauri::api::updater::UpdaterEvent;

tauri::updater::check_update().await.map(|update| {
    if update.is_some() {
        // Prompt user for update
        show_update_dialog();
    }
});
```

## Testing

### Unit Tests

```bash
# Run Rust tests
cargo test

# Run with logging
cargo test -- --nocapture

# Run specific test
cargo test audio_recording_test
```

### Integration Tests

```bash
# Run all tests
cargo test --test integration

# Test with UI
cargo tauri test
```

## Deployment

### GitHub Actions

```yaml
name: Build Desktop App

on:
  push:
    tags: [ 'v*' ]

jobs:
  build:
    strategy:
      matrix:
        platform: [macos-latest, ubuntu-latest, windows-latest]

    runs-on: ${{ matrix.platform }}

    steps:
    - uses: actions/checkout@v4

    - name: Install Rust
      uses: actions-rs/toolchain@v1
      with:
        toolchain: stable

    - name: Install Node.js
      uses: actions/setup-node@v4
      with:
        node-version: 20

    - name: Install dependencies
      run: npm install

    - name: Build Tauri app
      run: cargo tauri build

    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: ${{ matrix.platform }}-build
        path: src-tauri/target/release/bundle/
```

### Distribution

- **macOS**: App Store, direct download (.dmg)
- **Windows**: Microsoft Store, direct download (.msi)
- **Linux**: AppImage, Snap, Flatpak packages