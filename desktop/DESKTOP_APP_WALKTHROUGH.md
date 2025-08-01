# Desktop Application Walkthrough

## Table of Contents
- [Desktop Application Walkthrough](#desktop-application-walkthrough)
  - [Table of Contents](#table-of-contents)
  - [What is Electron?](#what-is-electron)
  - [Project Structure](#project-structure)
  - [Main Process vs Renderer Process](#main-process-vs-renderer-process)
    - [Main Process (`main.js`)](#main-process-mainjs)
    - [Renderer Process (`renderer/js/*.js`)](#renderer-process-rendererjsjs)
  - [Application Entry Point](#application-entry-point)
    - [1. Main Process Startup (`main.js`)](#1-main-process-startup-mainjs)
    - [2. Security Bridge (`preload.js`)](#2-security-bridge-preloadjs)
    - [3. HTML Structure (`index.html`)](#3-html-structure-indexhtml)
  - [Window Management](#window-management)
  - [Frontend Architecture](#frontend-architecture)
    - [1. Service Layer Pattern](#1-service-layer-pattern)
    - [2. Global Service Instances](#2-global-service-instances)
    - [3. Event-Driven Architecture](#3-event-driven-architecture)
  - [Authentication System](#authentication-system)
    - [How Authentication Works](#how-authentication-works)
    - [1. Login Process Flow](#1-login-process-flow)
    - [2. Token Management](#2-token-management)
    - [3. Session Persistence](#3-session-persistence)
  - [Chat System](#chat-system)
    - [Chat Architecture Overview](#chat-architecture-overview)
    - [1. Contact Loading Process](#1-contact-loading-process)
    - [2. Message Flow](#2-message-flow)
    - [3. UI State Management](#3-ui-state-management)
  - [WebSocket Communication](#websocket-communication)
    - [Real-time Features](#real-time-features)
    - [1. WebSocket Connection](#1-websocket-connection)
    - [2. Message Handling](#2-message-handling)
    - [3. Service Integration](#3-service-integration)
  - [Search Functionality](#search-functionality)
    - [Overview of Search Implementation](#overview-of-search-implementation)
    - [1. Search Flow](#1-search-flow)
    - [2. API Request with Conversation Context](#2-api-request-with-conversation-context)
    - [3. Backend Search Implementation](#3-backend-search-implementation)
    - [4. Search Results Display](#4-search-results-display)
    - [5. Search Result Interaction](#5-search-result-interaction)
  - [User Interface Components](#user-interface-components)
    - [Current User Display](#current-user-display)
    - [HTML Structure](#html-structure)
    - [CSS Styling](#css-styling)
  - [File Organization](#file-organization)
    - [Why This Structure?](#why-this-structure)
    - [1. Main Process Files](#1-main-process-files)
    - [2. Renderer Process Structure](#2-renderer-process-structure)
    - [3. Loading Order Importance](#3-loading-order-importance)
  - [How Everything Connects](#how-everything-connects)
    - [1. Application Startup Flow](#1-application-startup-flow)
    - [2. Service Communication Pattern](#2-service-communication-pattern)
    - [3. Data Flow Architecture](#3-data-flow-architecture)
    - [4. State Management](#4-state-management)
    - [5. Error Handling Strategy](#5-error-handling-strategy)
    - [6. Security Considerations](#6-security-considerations)
  - [Running the Application](#running-the-application)
    - [Development Mode](#development-mode)
    - [What Happens:](#what-happens)
    - [Production Build](#production-build)
  - [Key Takeaways](#key-takeaways)

## What is Electron?

Electron is a framework that allows you to build desktop applications using web technologies (HTML, CSS, JavaScript). It combines the Chromium rendering engine and the Node.js runtime, allowing you to build desktop apps with web technologies.

**Key Concepts:**
- **Main Process**: The Node.js process that manages application lifecycle and creates renderer processes
- **Renderer Process**: The process that displays the UI using Chromium (like a web browser)
- **IPC (Inter-Process Communication)**: How main and renderer processes communicate

## Project Structure

```
desktop/
├── main.js                 # Main process entry point
├── preload.js             # Security bridge between main and renderer
├── package.json           # Dependencies and app configuration
├── index.html            # Main UI structure
└── renderer/             # Frontend code (runs in renderer process)
    ├── css/
    │   ├── main.css      # Base styles and variables
    │   ├── components.css # UI component styles
    │   └── page.css      # Page-specific styles
    └── js/
        ├── app.js        # Main application logic
        ├── auth.js       # Authentication handling
        ├── chat.js       # Chat functionality
        ├── websocket.js  # Real-time communication
        ├── components.js # Reusable UI components
        ├── userStatus.js # User online/offline status
        ├── router.js     # Navigation management
        └── init.js       # Application initialization
```

## Main Process vs Renderer Process

### Main Process (`main.js`)
The main process is like the "server" of your desktop app. It:
- Creates and manages application windows
- Handles system-level operations (file system, notifications, etc.)
- Manages application lifecycle (startup, shutdown)
- Communicates with the operating system

```javascript
// main.js - Simplified version
const { app, BrowserWindow } = require('electron');

function createWindow() {
    const mainWindow = new BrowserWindow({
        width: 1200,
        height: 800,
        webPreferences: {
            preload: path.join(__dirname, 'preload.js'),
            contextIsolation: true,
            enableRemoteModule: false
        }
    });
    
    mainWindow.loadFile('index.html');
}

app.whenReady().then(createWindow);
```

### Renderer Process (`renderer/js/*.js`)
The renderer process is like the "client" - it's essentially a web page running in Chromium:
- Handles UI interactions
- Makes HTTP requests to your backend
- Manages application state
- Displays data to users

## Application Entry Point

### 1. Main Process Startup (`main.js`)
When you run `npm start`, Electron starts the main process:

```javascript
// 1. App becomes ready
app.whenReady().then(() => {
    createWindow();
});

// 2. Create window with security settings
function createWindow() {
    const mainWindow = new BrowserWindow({
        // Window configuration
        webPreferences: {
            preload: path.join(__dirname, 'preload.js'), // Security bridge
            contextIsolation: true,                       // Security isolation
            enableRemoteModule: false                     // Disable remote for security
        }
    });
    
    // 3. Load the HTML file
    mainWindow.loadFile('index.html');
}
```

### 2. Security Bridge (`preload.js`)
This file runs before the renderer process and safely exposes APIs:

```javascript
// preload.js - Safely expose APIs to renderer
const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
    setBadgeCount: (count) => ipcRenderer.invoke('set-badge-count', count),
    store: {
        get: (key) => ipcRenderer.invoke('store-get', key),
        set: (key, value) => ipcRenderer.invoke('store-set', key, value),
        delete: (key) => ipcRenderer.invoke('store-delete', key)
    }
});
```

### 3. HTML Structure (`index.html`)
The main UI structure loads all necessary files:

```html
<!-- Load CSS -->
<link rel="stylesheet" href="renderer/css/main.css">
<link rel="stylesheet" href="renderer/css/components.css">
<link rel="stylesheet" href="renderer/css/page.css">

<!-- UI Structure -->
<div id="app-container">
    <!-- Authentication, Chat UI, etc. -->
</div>

<!-- Load JavaScript in correct order -->
<script src="renderer/js/components.js"></script>
<script src="renderer/js/auth.js"></script>
<script src="renderer/js/websocket.js"></script>
<script src="renderer/js/userStatus.js"></script>
<script src="renderer/js/chat.js"></script>
<script src="renderer/js/router.js"></script>
<script src="renderer/js/app.js"></script>
<script src="renderer/js/init.js"></script>
```

## Window Management

The main process manages the application window:

```javascript
// Window creation with specific settings
const mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    icon: path.join(__dirname, 'assets/icon.png'),
    webPreferences: {
        preload: path.join(__dirname, 'preload.js'),
        contextIsolation: true,
        enableRemoteModule: false,
        nodeIntegration: false  // Security: disable node in renderer
    }
});
```

**Key Security Features:**
- `contextIsolation: true` - Isolates the main world from isolated world
- `nodeIntegration: false` - Prevents renderer from accessing Node.js APIs directly
- `preload.js` - Safe way to expose specific APIs to renderer

## Frontend Architecture

The frontend follows a modular architecture with clear separation of concerns:

### 1. Service Layer Pattern
Each major feature has its own service class:

```javascript
// auth.js - Authentication Service
class AuthService {
    constructor() {
        this.currentUser = null;
        this.token = null;
    }
    
    async login(email, password) { /* ... */ }
    async logout() { /* ... */ }
    async authenticatedFetch(url, options) { /* ... */ }
}

// chat.js - Chat Service  
class ChatService {
    constructor() {
        this.messages = {};
        this.contacts = [];
    }
    
    async loadContacts() { /* ... */ }
    async sendMessage(receiverId, content) { /* ... */ }
}
```

### 2. Global Service Instances
Services are created as global instances for easy access:

```javascript
// Create global instances
window.authService = new AuthService();
window.chatService = new ChatService();
window.websocketService = new WebSocketService();
```

### 3. Event-Driven Architecture
Services communicate through events and callbacks:

```javascript
// Chat service notifies listeners of changes
class ChatService {
    constructor() {
        this.listeners = new Set();
    }
    
    addListener(callback) {
        this.listeners.add(callback);
    }
    
    notifyListeners() {
        this.listeners.forEach(callback => callback(this.getState()));
    }
}
```

## Authentication System

### How Authentication Works

The authentication system manages user login/logout and maintains session state:

### 1. Login Process Flow

```javascript
// User clicks login button → triggers handleLogin
async handleLogin(e) {
    e.preventDefault();

    // 1. Get form data
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    // 2. Send to backend
    const response = await fetch('http://localhost:8080/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
    });

    // 3. Handle response
    if (response.ok) {
        const data = await response.json();
        this.setCurrentUser(data.user, data.token);
        this.showMainApp();
    }
}
```

### 2. Token Management

```javascript
class AuthService {
    setCurrentUser(user, token) {
        this.currentUser = user;
        this.token = token;

        // Store in Electron store for persistence
        await window.electronAPI.store.set('userData', user);
        await window.electronAPI.store.set('token', token);
    }

    // Add token to all API requests
    async authenticatedFetch(url, options = {}) {
        const headers = {
            'Authorization': `Bearer ${this.token}`,
            'Content-Type': 'application/json',
            ...options.headers
        };

        return fetch(`http://localhost:8080/${url}`, {
            ...options,
            headers
        });
    }
}
```

### 3. Session Persistence

```javascript
// On app startup, check for existing session
async checkExistingSession() {
    const token = await window.electronAPI.store.get('token');
    const userData = await window.electronAPI.store.get('userData');

    if (token && userData) {
        this.token = token;
        this.currentUser = userData;

        // Verify token is still valid
        const isValid = await this.verifyToken();
        if (isValid) {
            this.showMainApp();
        } else {
            this.logout(); // Clear invalid session
        }
    }
}
```

## Chat System

### Chat Architecture Overview

The chat system consists of several interconnected components:

1. **Contact Management** - Loading and displaying chat contacts
2. **Message Handling** - Sending, receiving, and displaying messages
3. **Real-time Updates** - WebSocket integration for live updates
4. **UI State Management** - Managing selected chats and UI updates

### 1. Contact Loading Process

```javascript
// ChatService loads contacts from backend
async loadContacts() {
    try {
        const response = await window.authService.authenticatedFetch('chat/contacts');
        const contacts = await response.json();

        this.contacts = contacts;
        this.notifyListeners(); // Update UI
    } catch (error) {
        console.error('Failed to load contacts:', error);
    }
}
```

### 2. Message Flow

```javascript
// Sending a message
async sendMessage(receiverId, content) {
    // 1. Create temporary message for immediate UI update
    const tempMessage = {
        id: `temp-${Date.now()}`,
        senderId: window.authService.currentUser.id,
        receiverId,
        content,
        createdAt: new Date(),
        isRead: false
    };

    // 2. Add to local state immediately
    if (!this.messages[receiverId]) {
        this.messages[receiverId] = [];
    }
    this.messages[receiverId].push(tempMessage);
    this.notifyListeners(); // Update UI immediately

    // 3. Send to server
    const response = await window.authService.authenticatedFetch('chat/send', {
        method: 'POST',
        body: JSON.stringify({ receiverId, content })
    });

    // 4. Replace temp message with real message
    const realMessage = await response.json();
    const index = this.messages[receiverId].findIndex(m => m.id === tempMessage.id);
    this.messages[receiverId][index] = realMessage;
    this.notifyListeners();
}
```

### 3. UI State Management

```javascript
// MessengerApp manages overall UI state
class MessengerApp {
    constructor() {
        this.selectedChat = null; // Currently selected conversation
        this.searchQuery = '';    // Contact search query
    }

    selectChat(contact) {
        this.selectedChat = contact;
        this.updateChatView();

        // Clear global search when switching chats
        const globalSearch = document.getElementById('global-search');
        if (globalSearch) {
            globalSearch.value = '';
        }
        this.clearSearchResults();

        // Load messages for this contact
        window.chatService.loadMessages(contact.userId);

        // Mark messages as read
        window.chatService.markMessagesAsRead(contact.userId);
    }

    updateChatView() {
        // Show/hide appropriate UI sections
        const noChatSelected = document.getElementById('no-chat-selected');
        const activeChat = document.getElementById('active-chat');

        if (this.selectedChat) {
            noChatSelected.style.display = 'none';
            activeChat.style.display = 'flex';
            this.updateChatHeader();
        } else {
            noChatSelected.style.display = 'flex';
            activeChat.style.display = 'none';
        }
    }
}
```

## WebSocket Communication

### Real-time Features

The WebSocket system enables real-time features:
- Instant message delivery
- Typing indicators
- Online/offline status
- Read receipts

### 1. WebSocket Connection

```javascript
class WebSocketService {
    constructor() {
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.subscribers = new Map();
    }

    connect() {
        if (!window.authService.token) return;

        this.ws = new WebSocket(`ws://localhost:8080/ws?token=${window.authService.token}`);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.attemptReconnect();
        };
    }
}
```

### 2. Message Handling

```javascript
handleMessage(message) {
    switch (message.type) {
        case 'new_message':
            this.notifySubscribers('message', message.data);
            break;
        case 'message_read':
            this.notifySubscribers('read', message.data);
            break;
        case 'typing':
            this.notifySubscribers('typing', message.data);
            break;
        case 'user_status':
            this.notifySubscribers('userStatus', message.data);
            break;
    }
}
```

### 3. Service Integration

```javascript
// Chat service subscribes to WebSocket events
initializeWebSocketSubscriptions() {
    // Subscribe to new messages
    this.messageUnsubscribe = window.websocketService.subscribe('message', (message) => {
        this.handleNewMessage(message);
    });

    // Subscribe to read receipts
    this.readUnsubscribe = window.websocketService.subscribe('read', (data) => {
        this.handleMessageRead(data);
    });

    // Subscribe to typing indicators
    this.typingUnsubscribe = window.websocketService.subscribe('typing', (data) => {
        this.handleTypingIndicator(data);
    });
}
```

## Search Functionality

### Overview of Search Implementation

The search functionality allows users to search for messages within their currently selected conversation. This was recently modified to be conversation-specific rather than global.

### 1. Search Flow

```javascript
// User types in search input → triggers handleGlobalSearch
async handleGlobalSearch(e) {
    const query = e.target.value.trim();

    // Clear existing timeout for debouncing
    if (this.searchTimeout) {
        clearTimeout(this.searchTimeout);
    }

    // Only search if query exists and chat is selected
    if (query.length < 1 || !this.selectedChat) {
        this.clearSearchResults();
        return;
    }

    // Debounce search requests (wait 300ms after user stops typing)
    this.searchTimeout = setTimeout(async () => {
        await this.performSearch(query);
    }, 300);
}
```

### 2. API Request with Conversation Context

```javascript
async performSearch(query) {
    try {
        // Include the selected chat user ID in the search
        const otherUserId = this.selectedChat.userId;
        const searchUrl = `chat/search?q=${encodeURIComponent(query)}&otherUserId=${encodeURIComponent(otherUserId)}&limit=20`;

        const response = await window.authService.authenticatedFetch(searchUrl);
        const data = await response.json();

        if (data.messages && data.messages.length > 0) {
            this.displaySearchResults(data.messages, query);
        } else {
            this.displaySearchResults([], query);
        }
    } catch (error) {
        console.error('Search error:', error);
        this.clearSearchResults();
    }
}
```

### 3. Backend Search Implementation

The backend receives the search request and filters messages:

```javascript
// Backend handler (Go)
func (h *Handler) SearchMessages(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    query := r.URL.Query().Get("q")
    otherUserID := r.URL.Query().Get("otherUserId") // New parameter

    // Search messages between userID and otherUserID
    messages, err := h.service.SearchMessages(userID, query, otherUserID, limit)
    // Return filtered results
}
```

### 4. Search Results Display

```javascript
displaySearchResults(messages, query) {
    const searchResults = document.getElementById('search-results');

    if (messages.length === 0) {
        searchResults.innerHTML = '<div class="no-results">No messages found</div>';
        return;
    }

    const resultsHTML = messages.map(message => {
        const isFromCurrentUser = message.senderId === window.authService.currentUser?.id;
        const senderName = isFromCurrentUser ? 'You' :
            `${message.sender?.firstName || 'Unknown'} ${message.sender?.lastName || 'User'}`.trim();

        // Highlight search term in message content
        const escapedQuery = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        const highlightedContent = message.content.replace(
            new RegExp(`(${escapedQuery})`, 'gi'),
            '<mark>$1</mark>'
        );

        return `
            <div class="search-result-item" data-message-id="${message.id}">
                <div class="search-result-contact">${senderName}</div>
                <div class="search-result-content">${highlightedContent}</div>
                <div class="search-result-time">${this.formatMessageTime(message.createdAt)}</div>
            </div>
        `;
    }).join('');

    searchResults.innerHTML = resultsHTML;
    searchResults.style.display = 'block';
}
```

### 5. Search Result Interaction

```javascript
// When user clicks a search result, scroll to that message
searchResults.querySelectorAll('.search-result-item').forEach(item => {
    item.addEventListener('click', () => {
        const messageId = item.dataset.messageId;
        this.clearSearchResults();

        // Scroll to the specific message
        setTimeout(() => {
            this.scrollToMessage(messageId);
        }, 100);
    });
});
```

## User Interface Components

### Current User Display

The application shows the current user's avatar and name in the header:

```javascript
// Update current user display when authentication state changes
updateCurrentUserDisplay(authState) {
    const currentUserInfo = document.getElementById('current-user-info');
    const currentUserAvatar = document.getElementById('current-user-avatar');

    if (!currentUserInfo || !currentUserAvatar) return;

    if (authState.isAuthenticated && authState.currentUser) {
        const user = authState.currentUser;
        const baseUrl = 'http://localhost:8080';

        // Set avatar source
        const avatarSrc = user.avatar
            ? `${baseUrl}/uploads/${user.avatar}`
            : './avatar.png';
        currentUserAvatar.src = avatarSrc;

        // Set tooltip with user's full name
        const fullName = `${user.firstName || ''} ${user.lastName || ''}`.trim() || user.email;
        currentUserAvatar.title = fullName;
        currentUserAvatar.alt = fullName;

        // Show the current user info
        currentUserInfo.style.display = 'flex';
    } else {
        // Hide the current user info when not authenticated
        currentUserInfo.style.display = 'none';
    }
}
```

### HTML Structure

```html
<!-- Header with current user info -->
<div class="header-right">
    <div class="current-user-info" id="current-user-info" style="display: none;">
        <img id="current-user-avatar" src="./avatar.png" alt="Current User" class="current-user-avatar" title="">
    </div>
    <button class="header-btn" id="logout-btn" title="Logout">
        <i class="fas fa-sign-out-alt"></i>
    </button>
</div>
```

### CSS Styling

```css
/* Current User Info */
.current-user-info {
    display: flex;
    align-items: center;
    margin-right: var(--spacing-sm);
}

.current-user-avatar {
    width: 36px;
    height: 36px;
    border-radius: var(--border-radius-full);
    cursor: pointer;
    transition: var(--transition-fast);
    border: 2px solid transparent;
    object-fit: cover;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.current-user-avatar:hover {
    border-color: var(--primary-color);
    transform: scale(1.05);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}
```

## File Organization

### Why This Structure?

The file organization follows separation of concerns and modularity principles:

### 1. Main Process Files
- `main.js` - Entry point, window management, system integration
- `preload.js` - Security bridge between main and renderer processes
- `package.json` - Dependencies, scripts, app metadata

### 2. Renderer Process Structure

```
renderer/
├── css/
│   ├── main.css      # Base styles, variables, and utilities
│   ├── components.css # UI component styles
│   └── page.css      # Page-specific styles and responsive design
└── js/
    ├── components.js # Reusable UI components and utilities
    ├── auth.js       # Authentication service
    ├── chat.js       # Chat functionality service
    ├── websocket.js  # Real-time communication
    ├── userStatus.js # User online/offline status
    ├── router.js     # Navigation and routing
    ├── app.js        # Main application controller
    └── init.js       # Application initialization
```

### 3. Loading Order Importance

The JavaScript files must be loaded in a specific order due to dependencies:

```html
<!-- 1. Base utilities and components -->
<script src="renderer/js/components.js"></script>

<!-- 2. Core services -->
<script src="renderer/js/auth.js"></script>
<script src="renderer/js/websocket.js"></script>
<script src="renderer/js/userStatus.js"></script>

<!-- 3. Feature services -->
<script src="renderer/js/chat.js"></script>
<script src="renderer/js/router.js"></script>

<!-- 4. Main application -->
<script src="renderer/js/app.js"></script>

<!-- 5. Initialization -->
<script src="renderer/js/init.js"></script>
```

**Why this order?**
- `components.js` provides utilities used by other files
- `auth.js` and `websocket.js` are core services needed by others
- `chat.js` depends on auth and websocket services
- `app.js` orchestrates all services
- `init.js` starts everything up

## How Everything Connects

### 1. Application Startup Flow

```
1. Electron starts main.js
   ↓
2. Main process creates window and loads index.html
   ↓
3. Renderer process loads JavaScript files in order
   ↓
4. init.js creates service instances and starts app
   ↓
5. App checks for existing session
   ↓
6. If authenticated: load chat data and connect WebSocket
   If not authenticated: show login form
```

### 2. Service Communication Pattern

```
User Action → App Controller → Service → Backend API
                ↓
            UI Update ← Service Listener ← Service Response
```

Example: Sending a message
```
1. User types message and presses Enter
2. MessengerApp.handleSendMessage() called
3. ChatService.sendMessage() called
4. HTTP request sent to backend
5. ChatService notifies listeners
6. MessengerApp updates UI
7. WebSocket receives confirmation
8. UI shows message as sent
```

### 3. Data Flow Architecture

```
Backend API ←→ HTTP Requests ←→ Services ←→ App Controller ←→ UI
     ↑                                                        ↓
WebSocket ←→ Real-time Updates ←→ Services ←→ Event Listeners ←→ UI Updates
```

### 4. State Management

Each service manages its own state:

```javascript
// AuthService state
{
    currentUser: { id, email, firstName, lastName, avatar },
    token: "jwt-token-string",
    isAuthenticated: true
}

// ChatService state
{
    contacts: [{ userId, firstName, lastName, avatar, isOnline }],
    messages: {
        "user-id-1": [message1, message2, ...],
        "user-id-2": [message3, message4, ...]
    },
    typingUsers: { "user-id": { timestamp, isTyping } },
    unreadCounts: { "user-id": 3 }
}

// MessengerApp state
{
    selectedChat: { userId, firstName, lastName },
    searchQuery: "search term"
}
```

### 5. Error Handling Strategy

```javascript
// Service level error handling
async sendMessage(receiverId, content) {
    try {
        const response = await this.authenticatedFetch('chat/send', {
            method: 'POST',
            body: JSON.stringify({ receiverId, content })
        });

        if (!response.ok) {
            throw new Error('Failed to send message');
        }

        return await response.json();
    } catch (error) {
        console.error('Send message error:', error);
        window.showToast('Failed to send message', 'error');
        throw error; // Re-throw for caller to handle
    }
}
```

### 6. Security Considerations

- **Context Isolation**: Renderer process can't access Node.js APIs directly
- **Preload Script**: Safe bridge for necessary APIs
- **Token Management**: JWT tokens stored in Electron store
- **HTTPS/WSS**: Secure communication with backend
- **Input Validation**: All user inputs are validated and sanitized

## Running the Application

### Development Mode
```bash
cd desktop
npm install
npm start
```

### What Happens:
1. Electron starts main process (`main.js`)
2. Main process creates window and loads `index.html`
3. Renderer process loads all JavaScript files
4. Application initializes and checks for existing session
5. User can login and start chatting

### Production Build
```bash
npm run build
```

This creates a packaged desktop application that can be distributed to users.

## Key Takeaways

1. **Electron = Web Tech + Desktop**: Use HTML/CSS/JS to build desktop apps
2. **Two Processes**: Main (Node.js) and Renderer (Chromium) with secure communication
3. **Service Architecture**: Modular services handle different concerns
4. **Event-Driven**: Services communicate through events and callbacks
5. **Real-time**: WebSocket integration for live features
6. **Security First**: Context isolation and secure API exposure
7. **State Management**: Each service manages its own state and notifies listeners
8. **Error Handling**: Graceful error handling at all levels
9. **User Experience**: Features like current user display, conversation-specific search, and real-time updates
10. **Maintainable Code**: Clear separation of concerns and modular architecture

This architecture provides a scalable, maintainable, and secure foundation for a desktop chat application with modern features and excellent user experience.
