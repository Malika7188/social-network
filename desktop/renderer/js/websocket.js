// WebSocket Service - Vanilla JS version of websocketService.js

const EVENT_TYPES = {
    POST_CREATED: 'post_created',
    POST_LIKED: 'post_liked',
    USER_STATS_UPDATED: 'user_stats_updated',
    USER_STATUS_UPDATE: 'user_status_update',
    PRIVATE_MESSAGE: 'private_message',
    MESSAGES_READ: 'messages_read',
    USER_TYPING: 'user_typing',
    NOTIFICATION_UPDATE: 'notification_Update'
};

class WebSocketService {
    constructor() {
        this.socket = null;
        this.listeners = {};
        this.reconnectTimeout = null;
        this.pingInterval = null;
        this.heartbeatInterval = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.baseReconnectDelay = 3000;
        this.tabId = `tab_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
        this.statusMessageTimeout = null;
        this.statusDebounceDelay = 500;
        this.lastStatus = null;
        this.isOnline = navigator.onLine;
        this.connectionQuality = 'good';
        this.lastPongTime = Date.now();

        // Bind methods
        this.connect = this.connect.bind(this);
        this.disconnect = this.disconnect.bind(this);
        this.subscribe = this.subscribe.bind(this);
        this.unsubscribe = this.unsubscribe.bind(this);
        this.sendMessage = this.sendMessage.bind(this);
        this.handleNetworkChange = this.handleNetworkChange.bind(this);
        this.checkConnectionQuality = this.checkConnectionQuality.bind(this);

        // Setup network event listeners
        this.setupNetworkListeners();
    }

    setupNetworkListeners() {
        // Listen for online/offline events
        window.addEventListener('online', this.handleNetworkChange);
        window.addEventListener('offline', this.handleNetworkChange);

        // Listen for connection type changes (if available)
        if ('connection' in navigator) {
            navigator.connection.addEventListener('change', this.handleNetworkChange);
        }
    }

    handleNetworkChange() {
        const wasOnline = this.isOnline;
        this.isOnline = navigator.onLine;

        console.log(`Network status changed: ${wasOnline ? 'online' : 'offline'} -> ${this.isOnline ? 'online' : 'offline'}`);

        if (!wasOnline && this.isOnline) {
            // Just came back online, attempt to reconnect
            console.log('Network restored, attempting to reconnect...');
            this.reconnectAttempts = 0; // Reset attempts on network recovery
            window.showToast('Network restored, reconnecting...', 'success');
            if (!this.isConnected()) {
                this.connect();
            }
        } else if (wasOnline && !this.isOnline) {
            // Just went offline
            console.log('Network lost');
            window.showToast('Network connection lost', 'error');
            this.updateConnectionStatus(false, 'network_offline');
            // Notify listeners immediately
            this.notifyListeners('connection', { connected: false, reason: 'network_offline' });
        }

        this.updateConnectionQuality();

        // Always notify listeners of network status change
        this.notifyListeners('network_change', {
            online: this.isOnline,
            wasOnline: wasOnline,
            connectionQuality: this.connectionQuality
        });
    }

    updateConnectionQuality() {
        if (!navigator.onLine) {
            this.connectionQuality = 'offline';
            return;
        }

        if ('connection' in navigator) {
            const connection = navigator.connection;
            const effectiveType = connection.effectiveType;

            if (effectiveType === '4g') {
                this.connectionQuality = 'good';
            } else if (effectiveType === '3g') {
                this.connectionQuality = 'fair';
            } else {
                this.connectionQuality = 'poor';
            }
        }
    }

    connect() {
        if (!navigator.onLine) {
            console.log('Cannot connect: device is offline');
            this.updateConnectionStatus(false, 'device_offline');
            this.notifyListeners('connection', { connected: false, reason: 'device_offline' });
            return;
        }

        // Show connecting state
        this.updateConnectingStatus();

        if (!window.authService || !window.authService.isAuthenticated()) {
            console.warn('Cannot connect WebSocket: not authenticated');
            return;
        }

        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            console.warn('WebSocket already connected');
            return;
        }

        if (this.socket && this.socket.readyState === WebSocket.CONNECTING) {
            console.log('WebSocket connection in progress');
            return;
        }

        try {
            const token = window.authService.token;
            const wsUrl = `ws://localhost:8080/ws?token=${token}&tabId=${this.tabId}&clientType=electron`;

            console.log('Connecting to WebSocket:', wsUrl);
            this.socket = new WebSocket(wsUrl);

            this.socket.onopen = this.handleOpen.bind(this);
            this.socket.onclose = this.handleClose.bind(this);
            this.socket.onerror = this.handleError.bind(this);
            this.socket.onmessage = this.handleMessage.bind(this);

        } catch (error) {
            console.error('WebSocket connection error:', error);
            this.scheduleReconnect();
        }
    }

    handleOpen() {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.lastPongTime = Date.now();
        window.showToast('Connected to server', 'success');

        // Update connection status
        this.updateConnectionStatus(true);

        // Start ping interval (every 30 seconds)
        if (this.pingInterval) clearInterval(this.pingInterval);
        this.pingInterval = setInterval(() => {
            if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify({ type: 'ping' }));
            }
        }, 30000);

        // Start heartbeat monitoring (check every 45 seconds)
        if (this.heartbeatInterval) clearInterval(this.heartbeatInterval);
        this.heartbeatInterval = setInterval(() => {
            this.checkConnectionQuality();
        }, 45000);

        // Notify listeners
        this.notifyListeners('connection', { connected: true, quality: this.connectionQuality });
    }

    handleClose(event) {
        console.log('WebSocket disconnected', event.code, event.reason);
        this.socket = null;

        // Clear intervals
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }

        // Handle different close codes
        if (event.code === 1006) {
            // Code 1006 can mean many things: server down, network issues, etc.
            // Don't automatically log out - just show connection lost message
            console.warn('WebSocket closed unexpectedly (server may be down)');

            // Only log out if we get a specific unauthorized message
            if (event.reason && event.reason.includes('unauthorized')) {
                console.warn('WebSocket closed due to unauthorized access');
                window.showToast('Session expired. Please log in again.', 'error');
                window.authService.logout(false);
                return;
            } else {
                // Server is likely down - show helpful message
                window.showToast('Server is unavailable. Your login is saved and will reconnect automatically.', 'warning');
            }
        } else if (event.code === 1008) {
            // Code 1008 is specifically for policy violations (like auth failures)
            console.warn('WebSocket closed due to authentication failure');
            window.showToast('Session expired. Please log in again.', 'error');
            window.authService.logout(false);
            return;
        }

        // Attempt reconnect if not normal closure and user is still authenticated
        if (event.code !== 1000 && window.authService && window.authService.isAuthenticated()) {
            if (navigator.onLine) {
                this.scheduleReconnect();
            } else {
                console.log('Not attempting reconnect: device is offline');
                this.notifyListeners('connection', { connected: false, reason: 'device_offline' });
            }
        }

        // Notify listeners and update status with reason
        this.updateConnectionStatus(false, 'server_disconnected');
        this.notifyListeners('connection', { connected: false, reason: 'server_disconnected' });
    }

    handleError(error) {
        console.error('WebSocket error:', error);
        if (this.socket && this.socket.readyState !== WebSocket.CLOSED) {
            this.socket.close();
        }
    }

    handleMessage(event) {
        try {
            const messages = event.data.split(/\n|\r\n/);

            for (const message of messages) {
                if (!message.trim()) continue;

                try {
                    const data = JSON.parse(message);

                    // Handle ping/pong
                    if (data.type === 'ping') {
                        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                            this.socket.send(JSON.stringify({ type: 'pong' }));
                        }
                        continue;
                    }

                    if (data.type === 'pong') {
                        this.lastPongTime = Date.now();
                        continue;
                    }

                    // Notify listeners of the event
                    this.notifyListeners(data.type, data.payload);

                } catch (innerError) {
                    console.error('Error parsing WebSocket message:', innerError, 'Raw:', message);
                }
            }
        } catch (error) {
            console.error('Error processing WebSocket message:', error, 'Raw:', event.data);
        }
    }

    scheduleReconnect() {
        if (!navigator.onLine) {
            console.log('Not scheduling reconnect: device is offline');
            return;
        }

        // Don't give up on reconnecting when server is down - keep trying
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.log('Maximum reconnection attempts reached, but continuing to retry...');
            // Reset attempts to continue trying, but with longer delays
            this.reconnectAttempts = this.maxReconnectAttempts - 1;
            window.showToast('Server appears to be down. Will keep trying to reconnect...', 'warning');
        }

        this.reconnectAttempts++;
        // Cap the delay at 30 seconds for server downtime scenarios
        const delay = Math.min(this.baseReconnectDelay * Math.pow(1.5, this.reconnectAttempts - 1), 30000);

        console.log(`Attempting to reconnect (${this.reconnectAttempts}) in ${delay / 1000} seconds`);

        this.reconnectTimeout = setTimeout(() => {
            if (navigator.onLine && window.authService && window.authService.isAuthenticated()) {
                this.connect();
            } else if (!navigator.onLine) {
                console.log('Skipping reconnect: device is offline');
            } else if (!window.authService || !window.authService.isAuthenticated()) {
                console.log('Skipping reconnect: user not authenticated');
            }
        }, delay);
    }

    checkConnectionQuality() {
        if (!this.isConnected()) {
            return;
        }

        const now = Date.now();
        const timeSinceLastPong = now - this.lastPongTime;

        // If we haven't received a pong in over 60 seconds, connection might be poor
        if (timeSinceLastPong > 60000) {
            console.warn('Poor connection detected - no pong received in', timeSinceLastPong / 1000, 'seconds');
            this.connectionQuality = 'poor';

            // If no pong for over 90 seconds, force reconnect
            if (timeSinceLastPong > 90000) {
                console.warn('Connection appears dead, forcing reconnect');
                this.socket.close();
            }
        }
    }

    disconnect() {
        // Clear timeouts
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }

        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }

        if (this.statusMessageTimeout) {
            clearTimeout(this.statusMessageTimeout);
            this.statusMessageTimeout = null;
        }

        // Close socket
        if (this.socket) {
            console.log('Closing WebSocket connection');
            this.socket.close();
            this.socket = null;
        }

        // Clear listeners
        this.listeners = {};
        this.reconnectAttempts = 0;
        this.updateConnectionStatus(false);
    }

    sendMessage(data) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(data));
            return true;
        }
        return false;
    }

    sendStatusMessage(statusType) {
        // Clear any pending status message
        if (this.statusMessageTimeout) {
            clearTimeout(this.statusMessageTimeout);
        }

        // Don't send if it's the same as the last status
        if (this.lastStatus === statusType) {
            return;
        }

        this.statusMessageTimeout = setTimeout(() => {
            if (this.sendMessage({ type: statusType })) {
                console.log(`Sent ${statusType} message`);
                this.lastStatus = statusType;
            }
            this.statusMessageTimeout = null;
        }, this.statusDebounceDelay);
    }

    subscribe(eventType, callback) {
        if (!this.listeners[eventType]) {
            this.listeners[eventType] = [];
        }
        this.listeners[eventType].push(callback);

        // Return unsubscribe function
        return () => this.unsubscribe(eventType, callback);
    }

    unsubscribe(eventType, callback) {
        if (this.listeners[eventType]) {
            this.listeners[eventType] = this.listeners[eventType].filter(cb => cb !== callback);
        }
    }

    notifyListeners(eventType, payload) {
        if (this.listeners[eventType]) {
            this.listeners[eventType].forEach(callback => {
                try {
                    callback(payload);
                } catch (error) {
                    console.error('Listener error:', error);
                }
            });
        }
    }

    isConnected() {
        return this.socket && this.socket.readyState === WebSocket.OPEN;
    }

    updateConnectionStatus(isConnected, reason = null) {
        const networkIndicator = document.getElementById('network-indicator');
        if (networkIndicator) {
            const statusText = networkIndicator.querySelector('span');
            const statusIcon = networkIndicator.querySelector('i');

            if (isConnected) {
                networkIndicator.classList.remove('offline', 'connecting', 'server-down');
                networkIndicator.classList.add('online');
                if (statusText) statusText.textContent = 'Connected';
                if (statusIcon) statusIcon.className = 'fas fa-wifi';
            } else {
                networkIndicator.classList.remove('online', 'connecting');

                // Show different states based on the reason
                if (reason === 'server_disconnected' && window.authService && window.authService.isAuthenticated()) {
                    networkIndicator.classList.remove('offline');
                    networkIndicator.classList.add('server-down');
                    if (statusText) statusText.textContent = 'Server unavailable';
                    if (statusIcon) statusIcon.className = 'fas fa-exclamation-triangle';
                } else {
                    networkIndicator.classList.remove('server-down');
                    networkIndicator.classList.add('offline');
                    if (statusText) statusText.textContent = 'Cannot connect to server';
                    if (statusIcon) statusIcon.className = 'fas fa-wifi-slash';
                }
            }
        }
    }

    updateConnectingStatus() {
        const networkIndicator = document.getElementById('network-indicator');
        if (networkIndicator) {
            const statusText = networkIndicator.querySelector('span');
            const statusIcon = networkIndicator.querySelector('i');

            networkIndicator.classList.remove('online', 'offline');
            networkIndicator.classList.add('connecting');
            if (statusText) statusText.textContent = 'Connecting...';
            if (statusIcon) statusIcon.className = 'fas fa-wifi';
        }
    }

    // Handle visibility changes
    handleVisibilityChange() {
        if (document.visibilityState === 'hidden') {
            this.sendStatusMessage('user_away');
        } else if (document.visibilityState === 'visible') {
            if (!this.isConnected() && window.authService && window.authService.isAuthenticated()) {
                this.connect();
            } else if (this.isConnected()) {
                this.sendStatusMessage('user_active');
            }
        }
    }
}

// Initialize the service
window.websocketService = new WebSocketService();

// Handle page visibility changes
document.addEventListener('visibilitychange', () => {
    window.websocketService.handleVisibilityChange();
});

// Export EVENT_TYPES for use in other modules
window.EVENT_TYPES = EVENT_TYPES;