# Social Network Platform

A modern, feature-rich social networking platform built with Next.js frontend, Electron desktop application, and Go backend, featuring real-time communications, group management, and comprehensive social features.

## Overview

This platform provides a complete social networking experience with features like user authentication, real-time chat, group management, post sharing, and more. Built with scalability and performance in mind, it uses Next.js for the web frontend, Electron for the cross-platform desktop application, and Go for the backend services. The desktop application seamlessly integrates with the web platform, allowing users to communicate across both platforms in real-time.

---

## Table of Contents

- [Social Network Platform](#social-network-platform)
  - [Overview](#overview)
  - [Table of Contents](#table-of-contents)
  - [🚀 Key Features](#-key-features)
    - [Authentication \& User Management](#authentication--user-management)
    - [Profile \& Social Features](#profile--social-features)
    - [Groups](#groups)
    - [Content Sharing](#content-sharing)
    - [Real-time Features](#real-time-features)
    - [Desktop Application](#desktop-application)
    - [File Management](#file-management)
  - [🛠 Technology Stack](#-technology-stack)
    - [Frontend](#frontend)
    - [Desktop Application](#desktop-application-1)
    - [Backend](#backend)
    - [DevOps \& Infrastructure](#devops--infrastructure)
  - [📁 Project Structure](#-project-structure)
    - [Frontend Architecture](#frontend-architecture)
    - [Desktop Application Architecture](#desktop-application-architecture)
    - [Backend Architecture](#backend-architecture)
  - [🚀 Getting Started](#-getting-started)
    - [Prerequisites](#prerequisites)
    - [Local Development Setup](#local-development-setup)
    - [Desktop Application Features](#desktop-application-features)
    - [Docker Deployment](#docker-deployment)
    - [Desktop Application Building](#desktop-application-building)
      - [Development Mode](#development-mode)
      - [Building for Distribution](#building-for-distribution)
      - [Desktop App Configuration](#desktop-app-configuration)
  - [📜 API Documentation](#-api-documentation)
    - [Authentication Endpoints](#authentication-endpoints)
    - [User \& Profile Endpoints](#user--profile-endpoints)
    - [Groups Endpoints](#groups-endpoints)
    - [Posts Endpoints](#posts-endpoints)
  - [Contributing](#contributing)
  - [License](#license)
  - [Acknowledgments](#acknowledgments)

---

## 🚀 Key Features

### Authentication & User Management
- Secure JWT-based authentication
- User registration and login
- Session management
- Real-time user status tracking

### Profile & Social Features
- Customizable user profiles with media uploads
- Follow/Unfollow functionality
- Activity feed
- Friend connections
- Privacy settings

### Groups
- Member management
- Group content sharing
- Invitation system
- Event organization

### Content Sharing
- Rich text posts
- Media upload support (images, videos)
- Privacy controls
- Comments and reactions

### Real-time Features
- WebSocket-powered live chat
- Instant notifications
- Online status indicators
- Live content updates

### Desktop Application
- **Cross-platform desktop app** built with Electron
- **Seamless integration** with web users
- **Native system features**: System tray, notifications, badges
- **Offline capabilities**: Local data storage and persistence
- **Auto-updates**: Built-in update mechanism
- **Enhanced security**: Context isolation and CSP

### File Management
- Secure file upload system
- Multiple media type support
- Organized storage structure
- Image processing

---

## 🛠 Technology Stack

### Frontend
- **Framework**: Next.js 13+ with App Router
- **State Management**: React Context API
- **Real-time Communication**: WebSocket
- **Styling**: CSS Modules
- **Development Tools**: ESLint, Prettier

### Desktop Application
- **Framework**: Electron
- **Architecture**: Main process (Node.js) + Renderer process (Chromium)
- **Security**: Context isolation, preload scripts, CSP
- **Storage**: Electron Store for local data persistence
- **Real-time**: WebSocket integration with backend
- **Cross-platform**: Windows, macOS, Linux support

### Backend
- **Language**: Go
- **Database**: SQLite
- **Authentication**: JWT
- **File Storage**: Local filesystem
- **API**: RESTful + WebSocket
- **Libraries**:
  - `gorilla/websocket` for WebSocket
  - `jwt-go` for authentication
  - Built-in SQLite support
  - Custom middleware

### DevOps & Infrastructure
- **Containerization**: Docker
- **Orchestration**: Docker Compose
- **Logging**: Structured logging with rotation
- **Development**: Hot-reloading for both frontend and backend

---

## 📁 Project Structure

### Frontend Architecture
```
frontend/
├── src/
│   ├── app/                # Next.js 13+ App Router
│   │   ├── (protected)/   # Protected routes
│   │   ├── register/      # Authentication pages
│   │   ├── layout.js      # Root layout
│   │   └── page.js        # Home page
│   ├── components/        # UI Components
│   │   ├── auth/         # Authentication components
│   │   ├── chat/         # Chat interface
│   │   ├── groups/       # Group management
│   │   ├── posts/        # Post creation/display
│   │   └── ui/           # Common UI elements
│   ├── context/          # React contexts
│   ├── services/         # API services
│   ├── styles/           # CSS modules
│   └── utils/            # Utility functions
```

### Desktop Application Architecture
```
desktop/
├── main.js              # Main Electron process
├── preload.js           # Security bridge
├── index.html           # Main HTML structure
├── package.json         # Dependencies and build config
├── assets/              # Icons and images
│   ├── icon.png         # App icon
│   ├── icon.ico         # Windows icon
│   └── icon.icns        # macOS icon
├── build/               # Build configuration
│   └── entitlements.mac.plist
└── renderer/            # Frontend code (Chromium)
    ├── css/
    │   ├── main.css     # Base styles and variables
    │   ├── components.css # UI component styles
    │   └── page.css     # Page-specific styles
    └── js/
        ├── app.js       # Main application logic
        ├── auth.js      # Authentication service
        ├── chat.js      # Chat functionality
        ├── websocket.js # Real-time communication
        ├── components.js # UI components
        ├── userStatus.js # User status management
        ├── router.js    # Navigation
        └── init.js      # Application initialization
```

### Backend Architecture
```
backend/
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   ├── auth/            # Authentication
│   ├── chat/            # Real-time chat
│   ├── event/           # Event handling
│   ├── follow/          # Follow system
│   ├── group/           # Group management
│   ├── post/            # Post handling
│   ├── profile/         # User profiles
│   └── websocket/       # WebSocket handling
├── pkg/
│   ├── db/              # Database operations
│   ├── filestore/       # File handling
│   ├── logger/          # Logging system
│   ├── middleware/      # HTTP middleware
│   └── models/          # Data models
└── data/
    ├── social_network.db # SQLite database
    └── uploads/         # User uploads
```

---

## 🚀 Getting Started

### Prerequisites

- Go 1.24 or higher
- Node.js 18 or higher
- Docker and Docker Compose (optional)
- Git

### Local Development Setup

1. **Clone the Repository**
```bash
git clone https://github.com/Athooh/social-network.git
cd social-network
```

2. **Backend Setup**
```bash
cd backend
go run cmd/api/main.go
```

3. **Frontend Setup (Web Application)**
```bash
cd frontend
npm install
npm run dev
```

4. **Desktop Application Setup**
```bash
cd desktop
npm install
npm start
```

5. **Environment Configuration**

Frontend (.env.local):
```
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
```

Backend (environment or .env):
```
PORT=8080
JWT_SECRET=your-secret-key
DB_PATH=./data/social_network.db
UPLOAD_DIR=./data/uploads
```

### Desktop Application Features

The desktop application provides:
- **Cross-platform compatibility**: Windows, macOS, and Linux
- **Real-time messaging**: Seamless communication with web users
- **System integration**: System tray, notifications, and badges
- **Offline storage**: Local data persistence with Electron Store
- **Auto-updates**: Built-in update mechanism
- **Security**: Context isolation and Content Security Policy

### Docker Deployment

1. **Build and Run with Docker Compose**
```bash
docker-compose up --build
```

The application will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

### Desktop Application Building

#### Development Mode
```bash
cd desktop
npm install
npm run dev        # Start in development mode
npm start          # Alternative start command
```

#### Building for Distribution

**Prerequisites for Building:**
- **Windows**: No additional tools needed
- **macOS**: Xcode command line tools
- **Linux**: Standard build tools

**Build Commands:**
```bash
cd desktop

# Build for current platform
npm run build

# Build for specific platforms
npm run build:win       # Windows (.exe installer + portable)
npm run build:mac       # macOS (.dmg + .zip)
npm run build:linux     # Linux (.AppImage + .deb)

# Build for all platforms (requires appropriate build environment)
npm run build:all

# Create unpacked directory (for testing)
npm run pack
```

**Build Output:**
Built applications will be in the `desktop/dist/` directory:
- **Windows**: `.exe` installer and portable version
- **macOS**: `.dmg` installer and `.zip` archive
- **Linux**: `.AppImage`, `.deb` packages

#### Desktop App Configuration

The desktop app connects to `http://localhost:8080` by default. Ensure the backend server is running before launching the desktop application.

**Auto-Updates:**
Configure the `publish` section in `desktop/package.json` to enable auto-updates via GitHub releases.

---

## 📜 API Documentation

### Authentication Endpoints
```
POST /api/auth/register     # User registration
POST /api/auth/login        # User login
POST /api/auth/logout       # User logout
GET  /api/auth/verify      # Verify JWT token
```

### User & Profile Endpoints
```
GET    /api/users/profile      # Get user profile
PUT    /api/users/profile      # Update profile
GET    /api/users/status       # Get online status
POST   /api/users/follow       # Follow user
DELETE /api/users/follow       # Unfollow user
```

### Groups Endpoints
```
POST   /api/groups            # Create group
GET    /api/groups            # List groups
GET    /api/groups/:id        # Get group details
PUT    /api/groups/:id        # Update group
DELETE /api/groups/:id        # Delete group
POST   /api/groups/:id/invite # Invite to group
```

### Posts Endpoints
```
POST   /api/posts            # Create post
GET    /api/posts            # List posts
GET    /api/posts/:id        # Get post details
PUT    /api/posts/:id        # Update post
DELETE /api/posts/:id        # Delete post
POST   /api/posts/:id/like   # Like post
```

## Contributing

We welcome contributions to the Social Network project! If you'd like to contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Commit your changes and push them to your fork.
4. Submit a pull request with a detailed description of your changes.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- **Next.js** for providing a powerful framework for building React applications.
- **Electron** for enabling cross-platform desktop application development with web technologies.
- **SQLite** for offering a lightweight and efficient database solution.
- **Docker** for simplifying containerization and deployment.

---

Thank you for checking out the Social Network project! If you have any questions or feedback, feel free to open an issue or reach out to the maintainers. Happy coding! 🚀
