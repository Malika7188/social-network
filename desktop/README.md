# Vibes Messenger Desktop App

A cross-platform desktop application for the Vibes social network, built with Electron.

## 🚀 Quick Start

### 👥 For Users

1. **Download** the appropriate package for your platform from releases
2. **Install** the application using the installer
3. **Start backend**: `cd backend && go run cmd/api/main.go`
4. **Launch** the desktop app - it will connect to the backend automatically

### 🛠️ For Developers

#### Setup
```bash
# Clone and install
git clone <your-repo>
cd vibes-social-network/desktop
npm install
```

#### Development
```bash
npm run dev     # Development mode
npm start       # Start the app
```

#### Building
```bash
npm run build           # Build for current platform
npm run build:win       # Windows
npm run build:mac       # macOS
npm run build:linux     # Linux
npm run build:all       # All platforms
```

## Features

- **Cross-platform**: Works on Windows, macOS, and Linux
- **Real-time messaging**: Connect with users on both desktop and web
- **System integration**: System tray, notifications, and badges
- **Auto-updates**: Automatic updates when new versions are available
- **Secure**: Content Security Policy and context isolation enabled
- **Offline support**: Local data storage with Electron Store

## Architecture

The desktop app is a thin client that connects to the Go backend server:

```
Desktop App (Electron) ←→ Backend Server (Go) ←→ Database (SQLite)
                                ↕
                        Web App (Next.js)
```

Both desktop and web users can communicate with each other through the shared backend.

## Building

### Prerequisites

- Node.js 18+
- Platform-specific build tools:
  - **Windows**: No additional tools needed
  - **macOS**: Xcode command line tools
  - **Linux**: Standard build tools

### Build Commands

```bash
# Install dependencies
npm install

# Create icons (optional)
npm run create-icons

# Build for current platform
npm run build

# Build with helper script (recommended)
npm run build-app

# Build for specific platforms
npm run build:win
npm run build:mac
npm run build:linux
npm run build:all
```

### Output

Built applications will be in the `dist/` directory:

- **Windows**: `.exe` installer and portable version
- **macOS**: `.dmg` installer and `.zip` archive
- **Linux**: `.AppImage`, `.deb`, and `.rpm` packages

## Configuration

### Backend Connection

The app connects to `http://localhost:8080` by default. The backend server must be running for the app to function.

### Auto-Updates

Configure the `publish` section in `package.json` to enable auto-updates via GitHub releases.

## Development

### Project Structure

```
desktop/
├── main.js              # Main Electron process
├── preload.js           # Preload script for security
├── index.html           # Main HTML file
├── renderer/            # Renderer process files
│   ├── css/            # Stylesheets
│   └── js/             # JavaScript modules
├── assets/             # Icons and images
├── scripts/            # Build and utility scripts
└── build/              # Build configuration files
```

### Security

- Node.js integration disabled in renderer
- Context isolation enabled
- Content Security Policy enforced
- External URLs filtered through main process

## Troubleshooting

### Common Issues

1. **App won't start**: Check that the backend is running on port 8080
2. **Build fails**: Ensure you have the required build tools installed
3. **Icons missing**: Run `npm run create-icons` to generate placeholder icons
4. **Connection issues**: Verify firewall settings allow localhost:8080

### Getting Help

- Check the [PACKAGING.md](PACKAGING.md) guide for detailed build instructions
- Review the console logs for error messages
- Ensure the backend server is accessible at `http://localhost:8080`

## 🌟 Features

- **Cross-platform**: Windows, macOS, Linux support
- **Real-time sync**: Messages sync between desktop and browser users
- **System integration**: Tray icon, notifications, badges
- **Offline storage**: Local data persistence
- **Auto-updates**: Built-in update mechanism
- **Security**: CSP, context isolation, sandboxed renderer

## 🎯 Backend Integration

The desktop app connects to your backend server on `localhost:8080`. Make sure to start the backend before launching the desktop app.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test on multiple platforms if possible
5. Submit a pull request

## License

This project is licensed under the ISC License.
