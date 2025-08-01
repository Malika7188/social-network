const { contextBridge, ipcRenderer } = require('electron');

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld('electronAPI', {
    // Store methods
    store: {
        get: (key) => ipcRenderer.invoke('store-get', key),
        set: (key, value) => ipcRenderer.invoke('store-set', key, value),
        delete: (key) => ipcRenderer.invoke('store-delete', key)
    },

    // Cookie methods
    clearCookies: () => ipcRenderer.invoke('clear-cookies'),

    // External URL opening (secured through main process)
    openExternal: (url) => ipcRenderer.invoke('open-external', url),

    // Notification methods
    showNotification: (options) => ipcRenderer.invoke('show-notification', options),
    setBadgeCount: (count) => ipcRenderer.invoke('set-badge-count', count),

    // Network status
    checkOnlineStatus: () => ipcRenderer.invoke('check-online-status'),

    // Platform info
    platform: process.platform,

    // App info
    versions: process.versions
});

// Network status monitoring
window.addEventListener('DOMContentLoaded', () => {
    const updateOnlineStatus = () => {
        const statusElement = document.getElementById('connection-status');
        if (statusElement) {
            if (navigator.onLine) {
                statusElement.classList.remove('offline');
                statusElement.classList.add('online');
                statusElement.textContent = 'Online';
            } else {
                statusElement.classList.remove('online');
                statusElement.classList.add('offline');
                statusElement.textContent = 'Offline';
            }
        }
    };

    window.addEventListener('online', updateOnlineStatus);
    window.addEventListener('offline', updateOnlineStatus);
    updateOnlineStatus();
});