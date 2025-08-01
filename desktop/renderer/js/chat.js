// Chat Service - Vanilla JS version of chatService.js

class ChatService {
    constructor() {
        this.messages = {};
        this.typingUsers = {};
        this.unreadCounts = {};
        this.contacts = [];
        this.isInitialized = false;
        this.listeners = new Set();

        // Bind methods
        this.init = this.init.bind(this);
        this.loadContacts = this.loadContacts.bind(this);
        this.loadMessages = this.loadMessages.bind(this);
        this.sendMessage = this.sendMessage.bind(this);
        this.markMessagesAsRead = this.markMessagesAsRead.bind(this);
        this.sendTypingIndicator = this.sendTypingIndicator.bind(this);
    }

    init() {
        if (this.isInitialized) return;

        if (!window.websocketService || !window.authService) {
            console.warn('Dependencies not ready for ChatService initialization');
            return;
        }

        this.initializeWebSocketSubscriptions();
        this.isInitialized = true;
    }

    initializeWebSocketSubscriptions() {
        if (!window.authService.currentUser?.id) return;

        // Handle new messages
        this.messageUnsubscribe = window.websocketService.subscribe(
            window.EVENT_TYPES.PRIVATE_MESSAGE,
            (payload) => this.handleNewMessage(payload)
        );

        // Handle messages read
        this.readUnsubscribe = window.websocketService.subscribe(
            window.EVENT_TYPES.MESSAGES_READ,
            (payload) => this.handleMessagesRead(payload)
        );

        // Handle typing indicators
        this.typingUnsubscribe = window.websocketService.subscribe(
            window.EVENT_TYPES.USER_TYPING,
            (payload) => this.handleTypingIndicator(payload)
        );

        console.log('Chat WebSocket subscriptions initialized');
    }

    handleNewMessage(payload) {
        if (!payload) return;

        const {
            senderId,
            receiverId,
            content,
            createdAt,
            messageId,
            isRead,
            senderName,
            senderAvatar
        } = payload;

        const currentUserId = window.authService.currentUser.id;
        const contactId = currentUserId === senderId ? receiverId : senderId;

        // Create message object
        const message = {
            id: messageId,
            senderId,
            receiverId,
            content,
            createdAt,
            isRead,
            sender: {
                id: senderId,
                firstName: senderName?.split(' ')[0] || '',
                lastName: senderName?.split(' ')[1] || '',
                avatar: senderAvatar
            }
        };

        // Update messages state
        if (!this.messages[contactId]) {
            this.messages[contactId] = [];
        }

        const contactMessages = this.messages[contactId];

        // Check if message already exists to avoid duplicates
        const messageExists = contactMessages.some(msg =>
            msg.id === messageId ||
            (msg.isTemp && msg.senderId === senderId && msg.content === content)
        );

        if (!messageExists) {
            this.messages[contactId].push(message);
        } else {
            // Replace temp message with server version
            this.messages[contactId] = contactMessages.map(msg =>
                msg.isTemp && msg.senderId === senderId && msg.content === content
                    ? message
                    : msg
            );
        }

        // Update unread counts if we're the receiver
        if (receiverId === currentUserId && !isRead) {
            this.unreadCounts[senderId] = (this.unreadCounts[senderId] || 0) + 1;

            // Show desktop notification
            this.showMessageNotification(message, senderName);
        }

        // Update contacts list with latest message
        this.updateContactWithLatestMessage(contactId, message);

        this.notifyListeners();
    }

    handleMessagesRead(payload) {
        if (!payload) return;

        const { senderId, receiverId, readAt } = payload;
        const currentUserId = window.authService.currentUser.id;

        // If we're the sender, update our messages to the receiver as read
        if (senderId === currentUserId) {
            const contactMessages = this.messages[receiverId] || [];
            this.messages[receiverId] = contactMessages.map(msg => {
                if (msg.senderId === currentUserId && !msg.isRead) {
                    return { ...msg, isRead: true, readAt };
                }
                return msg;
            });
        }

        // If we're the receiver, update our unread count
        if (receiverId === currentUserId) {
            this.unreadCounts[senderId] = 0;
        }

        this.notifyListeners();
    }

    handleTypingIndicator(payload) {
        if (!payload) return;

        const { senderId, timestamp } = payload;
        const timestampNum = parseInt(timestamp);

        // Clear any existing timeout for this user
        if (this.typingUsers[senderId] && this.typingUsers[senderId].timeout) {
            clearTimeout(this.typingUsers[senderId].timeout);
        }

        // Set typing status with timestamp and timeout
        const timeout = setTimeout(() => {
            if (this.typingUsers[senderId] && this.typingUsers[senderId].timestamp === timestampNum) {
                delete this.typingUsers[senderId];
                this.notifyListeners();
            }
        }, 2000); // Reduced from 3 seconds to 2 seconds for faster response

        this.typingUsers[senderId] = {
            timestamp: timestampNum,
            timeout: timeout
        };

        this.notifyListeners();
    }

    async loadContacts() {
        try {
            const response = await window.authService.authenticatedFetch('chat/contacts');
            if (!response.ok) throw new Error('Failed to load contacts');

            const data = await response.json();
            this.contacts = data || [];

            this.notifyListeners();
            return this.contacts;
        } catch (error) {
            console.error('Error loading contacts:', error);
            window.showToast('Failed to load contacts', 'error');
            return [];
        }
    }

    async loadMessages(contactId, limit = 100, offset = 0) {
        try {
            const response = await window.authService.authenticatedFetch(
                `chat/messages?userId=${contactId}&limit=${limit}&offset=${offset}`
            );

            if (!response.ok) throw new Error('Failed to load messages');

            const data = await response.json();
            this.messages[contactId] = data || [];

            this.notifyListeners();
            return data;
        } catch (error) {
            console.error('Error loading messages:', error);
            window.showToast('Failed to load messages', 'error');
            return [];
        }
    }

    async sendMessage(receiverId, content) {
        try {
            // Add temporary message to UI immediately
            const tempMessage = {
                id: `temp_${Date.now()}`,
                senderId: window.authService.currentUser.id,
                receiverId,
                content,
                createdAt: new Date().toISOString(),
                isRead: false,
                isTemp: true,
                sender: {
                    id: window.authService.currentUser.id,
                    firstName: window.authService.currentUser.firstName,
                    lastName: window.authService.currentUser.lastName,
                    avatar: window.authService.currentUser.avatar
                }
            };

            if (!this.messages[receiverId]) {
                this.messages[receiverId] = [];
            }
            this.messages[receiverId].push(tempMessage);
            this.notifyListeners();

            // Send to server
            const response = await window.authService.authenticatedFetch('chat/send', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    receiverId,
                    content
                })
            });

            if (!response.ok) {
                // Remove temp message on error
                this.messages[receiverId] = this.messages[receiverId].filter(
                    msg => msg.id !== tempMessage.id
                );
                this.notifyListeners();
                throw new Error('Failed to send message');
            }

            const message = await response.json();
            return message;
        } catch (error) {
            console.error('Error sending message:', error);
            window.showToast('Failed to send message', 'error');
            throw error;
        }
    }

    async markMessagesAsRead(senderId) {
        try {
            const response = await window.authService.authenticatedFetch('chat/mark-read', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    senderId
                })
            });

            if (!response.ok) throw new Error('Failed to mark messages as read');

            // Update local state
            this.unreadCounts[senderId] = 0;

            // Update message read status locally
            const senderMessages = this.messages[senderId] || [];
            this.messages[senderId] = senderMessages.map(msg => {
                if (msg.senderId === senderId && !msg.isRead) {
                    return { ...msg, isRead: true, readAt: new Date().toISOString() };
                }
                return msg;
            });

            this.notifyListeners();
            return true;
        } catch (error) {
            console.error('Error marking messages as read:', error);
            return false;
        }
    }

    async sendTypingIndicator(receiverId) {
        try {
            await window.authService.authenticatedFetch('chat/typing', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    receiverId
                })
            });
            return true;
        } catch (error) {
            console.error('Error sending typing indicator:', error);
            return false;
        }
    }

    updateContactWithLatestMessage(contactId, message) {
        const contactIndex = this.contacts.findIndex(c => c.userId === contactId);
        if (contactIndex !== -1) {
            this.contacts[contactIndex] = {
                ...this.contacts[contactIndex],
                lastMessage: message.content,
                lastMessageSenderId: message.senderId,
                lastSent: message.createdAt,
                unreadCount: this.unreadCounts[contactId] || 0
            };
        }
    }

    async showMessageNotification(message, senderName) {
        try {
            // Check if window is focused
            if (document.hasFocus()) return;

            await window.electronAPI.showNotification({
                title: senderName || 'New Message',
                body: message.content,
                sound: true
            });

            // Update badge count
            const totalUnread = Object.values(this.unreadCounts).reduce((sum, count) => sum + count, 0);
            await window.electronAPI.setBadgeCount(totalUnread);
        } catch (error) {
            console.error('Error showing notification:', error);
        }
    }

    // Event listener management
    addListener(callback) {
        this.listeners.add(callback);
        return () => this.listeners.delete(callback);
    }

    notifyListeners() {
        this.listeners.forEach(callback => {
            try {
                callback({
                    messages: this.messages,
                    typingUsers: this.typingUsers,
                    unreadCounts: this.unreadCounts,
                    contacts: this.contacts,
                    isInitialized: this.isInitialized
                });
            } catch (error) {
                console.error('Listener error:', error);
            }
        });
    }

    cleanup() {
        if (this.messageUnsubscribe) this.messageUnsubscribe();
        if (this.readUnsubscribe) this.readUnsubscribe();
        if (this.typingUnsubscribe) this.typingUnsubscribe();

        this.isInitialized = false;
        this.listeners.clear();
    }
}

// Create global instance
window.chatService = new ChatService();