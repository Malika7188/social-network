// User Status Service - Vanilla JS version of userStatusService.js

class UserStatusService {
    constructor() {
        this.onlineUsers = {};
        this.listeners = new Set();
        this.isInitialized = false;
        this.recentNotifications = new Map(); // Track recent notifications to prevent spam
        this.notificationCooldown = 5000; // 5 seconds cooldown between notifications for same user

        // Bind methods
        this.init = this.init.bind(this);
        this.setUserStatus = this.setUserStatus.bind(this);
        this.isUserOnline = this.isUserOnline.bind(this);
        this.handleUserStatusUpdate = this.handleUserStatusUpdate.bind(this);
        this.initializeStatuses = this.initializeStatuses.bind(this);
        this.addListener = this.addListener.bind(this);
        this.removeListener = this.removeListener.bind(this);
        this.notifyListeners = this.notifyListeners.bind(this);
    }

    init() {
        if (this.isInitialized) return;

        // Subscribe to user status updates from WebSocket
        this.statusUnsubscribe = window.websocketService.subscribe(
            window.EVENT_TYPES.USER_STATUS_UPDATE,
            this.handleUserStatusUpdate
        );

        this.isInitialized = true;
        console.log('User Status Service initialized');
    }

    handleUserStatusUpdate(payload) {
        if (!payload || !payload.userId) {
            console.warn('Invalid user status update payload:', payload);
            return;
        }

        const { userId, isOnline } = payload;
        console.log(`User status update: ${userId} is now ${isOnline ? 'online' : 'offline'}`);

        // Check if this is actually a status change
        const previousStatus = this.onlineUsers[userId];
        if (previousStatus === isOnline) {
            console.log(`Status unchanged for ${userId}, skipping notification`);
            return;
        }

        // Update the status
        this.setUserStatus(userId, isOnline);

        // Show toast notification for status changes (with cooldown)
        if (window.authService.currentUser && userId !== window.authService.currentUser.id) {
            this.showStatusNotification(userId, isOnline);
        }
    }

    showStatusNotification(userId, isOnline) {
        const now = Date.now();
        const lastNotification = this.recentNotifications.get(userId);

        // Check if we're in cooldown period
        if (lastNotification && (now - lastNotification) < this.notificationCooldown) {
            console.log(`Notification cooldown active for ${userId}, skipping toast`);
            return;
        }

        // Show the notification
        const user = this.getUserDisplayName(userId);
        const statusText = isOnline ? 'came online' : 'went offline';
        window.showToast(`${user} ${statusText}`, isOnline ? 'success' : 'info', 4000);

        // Update the last notification time
        this.recentNotifications.set(userId, now);

        // Clean up old entries (older than cooldown period)
        this.cleanupNotificationHistory();
    }

    cleanupNotificationHistory() {
        const now = Date.now();
        for (const [userId, timestamp] of this.recentNotifications.entries()) {
            if (now - timestamp > this.notificationCooldown) {
                this.recentNotifications.delete(userId);
            }
        }
    }

    setUserStatus(userId, isOnline) {
        if (!userId) return;

        const wasOnline = this.onlineUsers[userId];
        this.onlineUsers[userId] = isOnline;

        // Only notify if status actually changed
        if (wasOnline !== isOnline) {
            this.notifyListeners();
        }
    }

    isUserOnline(userId, defaultStatus = false) {
        if (!userId) return defaultStatus;

        // If we have a status update from WebSocket, use that
        if (this.onlineUsers[userId] !== undefined) {
            return this.onlineUsers[userId];
        }

        // Otherwise use the provided default status (from API)
        return defaultStatus;
    }

    initializeStatuses(contacts) {
        if (!contacts || !Array.isArray(contacts)) return;

        console.log('Initializing user statuses from contacts:', contacts.length);

        const newStatuses = {};
        contacts.forEach(contact => {
            if (contact.userId && contact.isOnline !== undefined) {
                newStatuses[contact.userId] = contact.isOnline;
            }
        });

        // Update all statuses at once
        this.onlineUsers = { ...this.onlineUsers, ...newStatuses };
        this.notifyListeners();
    }

    clearUserStatus(userId) {
        if (!userId) return;
        this.setUserStatus(userId, false);
    }

    getUserDisplayName(userId) {
        // Try to get user name from chat contacts
        if (window.chatService && window.chatService.contacts) {
            const contact = window.chatService.contacts.find(c => c.userId === userId);
            if (contact) {
                return `${contact.firstName} ${contact.lastName}`.trim();
            }
        }

        // Fallback to user ID
        return `User ${userId.substring(0, 8)}`;
    }

    addListener(callback) {
        this.listeners.add(callback);
        
        // Return unsubscribe function
        return () => this.removeListener(callback);
    }

    removeListener(callback) {
        this.listeners.delete(callback);
    }

    notifyListeners() {
        this.listeners.forEach(callback => {
            try {
                callback({
                    onlineUsers: this.onlineUsers
                });
            } catch (error) {
                console.error('User status listener error:', error);
            }
        });
    }

    // Get all online users
    getOnlineUsers() {
        return { ...this.onlineUsers };
    }

    // Get online status for multiple users
    getUsersStatus(userIds) {
        const statuses = {};
        userIds.forEach(userId => {
            statuses[userId] = this.isUserOnline(userId);
        });
        return statuses;
    }

    cleanup() {
        if (this.statusUnsubscribe) {
            this.statusUnsubscribe();
        }
        
        this.listeners.clear();
        this.onlineUsers = {};
        this.isInitialized = false;
    }
}

// Initialize the service
window.userStatusService = new UserStatusService();

// Export for use in other modules
window.UserStatusService = UserStatusService;
