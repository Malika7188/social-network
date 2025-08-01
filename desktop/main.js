const { app, BrowserWindow, ipcMain, Menu, Tray, Notification, shell } = require('electron');
const { autoUpdater } = require('electron-updater');
const Store = require('electron-store');
const path = require('path');

// Initialize electron store for persistent data
const store = new Store();

let mainWindow;
let tray = null;

// Enable live reload for development
if (process.env.NODE_ENV === 'development') {
    require('electron-reload')(__dirname, {
        electron: path.join(__dirname, '..', 'node_modules', '.bin', 'electron'),
        hardResetMethod: 'exit'
    });
}

function createWindow() {
    // Create the browser window
    mainWindow = new BrowserWindow({
        width: 1200,
        height: 800,
        minWidth: 800,
        minHeight: 600,
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            enableRemoteModule: false,
            preload: path.join(__dirname, 'preload.js'),
            webSecurity: true,
            allowRunningInsecureContent: false
        },
        icon: path.join(__dirname, 'assets', 'icon.png'), // Add your app icon
        show: false, // Don't show until ready
        titleBarStyle: process.platform === 'darwin' ? 'hiddenInset' : 'default'
    });

    // Set up session to handle cookies properly
    const session = mainWindow.webContents.session;

    // Enable cookie support for your domain
    session.webRequest.onBeforeSendHeaders((details, callback) => {
        // Add any additional headers if needed
        callback({ requestHeaders: details.requestHeaders });
    });

    // Load the app
    mainWindow.loadFile('index.html');

    // Show window when ready to prevent visual flash
    mainWindow.once('ready-to-show', () => {
        mainWindow.show();

        // Focus window if already logged in
        const userData = store.get('userData');
        if (userData) {
            mainWindow.focus();
        }
    });

    // Handle window closed
    mainWindow.on('closed', () => {
        mainWindow = null;
    });

    // Handle minimize to tray (Windows/Linux)
    mainWindow.on('minimize', (event) => {
        if (process.platform !== 'darwin') {
            event.preventDefault();
            mainWindow.hide();
        }
    });

    // Handle close button
    mainWindow.on('close', (event) => {
        if (process.platform === 'darwin') {
            event.preventDefault();
            mainWindow.hide();
        }
    });

    // Open external links in default browser
    mainWindow.webContents.setWindowOpenHandler(({ url }) => {
        require('electron').shell.openExternal(url);
        return { action: 'deny' };
    });
}

function createTray() {
    const iconPath = path.join(__dirname, 'assets', 'tray-icon.png');
    tray = new Tray(iconPath);

    const contextMenu = Menu.buildFromTemplate([
        {
            label: 'Show NoteBook Messenger',
            click: () => {
                mainWindow.show();
                mainWindow.focus();
            }
        },
        {
            label: 'Quit',
            click: () => {
                app.isQuitting = true;
                app.quit();
            }
        }
    ]);

    tray.setToolTip('NoteBook Messenger');
    tray.setContextMenu(contextMenu);

    // Show window on tray click
    tray.on('click', () => {
        if (mainWindow.isVisible()) {
            mainWindow.hide();
        } else {
            mainWindow.show();
            mainWindow.focus();
        }
    });
}

// App event listeners
app.whenReady().then(() => {
    createWindow();
    createTray();

    // Check for updates
    if (process.env.NODE_ENV === 'production') {
        autoUpdater.checkForUpdatesAndNotify();
    }
});

app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
        app.quit();
    }
});

app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
        createWindow();
    } else {
        mainWindow.show();
    }
});

app.on('before-quit', () => {
    app.isQuitting = true;
});

// IPC Handlers
ipcMain.handle('store-get', (event, key) => {
    return store.get(key);
});

ipcMain.handle('store-set', (event, key, value) => {
    store.set(key, value);
});

ipcMain.handle('store-delete', (event, key) => {
    store.delete(key);
});

ipcMain.handle('show-notification', (event, options) => {
    if (Notification.isSupported()) {
        const notification = new Notification({
            title: options.title || 'NoteBook Messenger',
            body: options.body || '',
            icon: options.icon || path.join(__dirname, 'assets', 'icon.png'),
            sound: options.sound !== false
        });

        notification.on('click', () => {
            mainWindow.show();
            mainWindow.focus();
        });

        notification.show();
        return true;
    }
    return false;
});

ipcMain.handle('set-badge-count', (event, count) => {
    if (process.platform === 'darwin') {
        app.dock.setBadge(count > 0 ? count.toString() : '');
    }
});

ipcMain.handle('open-external', (event, url) => {
    // Security check - only allow specific domains
    const allowedDomains = ['localhost:3000', 'localhost:8080'];
    const urlObj = new URL(url);

    if (allowedDomains.some(domain => urlObj.host === domain)) {
        return shell.openExternal(url);
    } else {
        console.warn('Blocked external URL:', url);
        return false;
    }
});

ipcMain.handle('clear-cookies', async () => {
    if (mainWindow) {
        const session = mainWindow.webContents.session;
        await session.clearStorageData({
            storages: ['cookies']
        });
        return true;
    }
    return false;
});

ipcMain.handle('check-online-status', () => {
    return mainWindow.webContents.session.netLog ? true : false;
});



// Auto-updater events
autoUpdater.on('checking-for-update', () => {
    console.log('Checking for update...');
});

autoUpdater.on('update-available', (info) => {
    console.log('Update available.');
});

autoUpdater.on('update-not-available', (info) => {
    console.log('Update not available.');
});

autoUpdater.on('error', (err) => {
    console.log('Error in auto-updater. ' + err);
});

autoUpdater.on('download-progress', (progressObj) => {
    let log_message = "Download speed: " + progressObj.bytesPerSecond;
    log_message = log_message + ' - Downloaded ' + progressObj.percent + '%';
    log_message = log_message + ' (' + progressObj.transferred + "/" + progressObj.total + ')';
    console.log(log_message);
});

autoUpdater.on('update-downloaded', (info) => {
    console.log('Update downloaded');
    autoUpdater.quitAndInstall();
});