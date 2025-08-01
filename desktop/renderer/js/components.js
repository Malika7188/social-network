// Utility Functions and Components - Vanilla JS versions

// Toast Notification System
class ToastManager {
    constructor() {
        this.toasts = [];
        this.container = document.getElementById('toast-container');
        this.counter = 0;
    }

    show(message, type = 'info', duration = 3000) {
        const id = `toast-${Date.now()}-${this.counter++}`;
        const toast = this.createToast(id, message, type, duration);

        this.toasts.push({ id, element: toast, timeout: null });
        this.container.appendChild(toast);

        // Auto-remove after duration
        const timeout = setTimeout(() => {
            this.remove(id);
        }, duration);

        // Update timeout reference
        const toastData = this.toasts.find(t => t.id === id);
        if (toastData) {
            toastData.timeout = timeout;
        }

        return id;
    }

    createToast(id, message, type, duration) {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.dataset.id = id;

        const icon = this.getIcon(type);

        toast.innerHTML = `
            <div class="toast-icon">
                <i class="${icon}"></i>
            </div>
            <div class="toast-message">${message}</div>
            <button class="toast-close" onclick="window.toastManager.remove('${id}')">
                <i class="fas fa-times"></i>
            </button>
        `;

        return toast;
    }

    getIcon(type) {
        switch (type) {
            case 'success': return 'fas fa-check-circle';
            case 'error': return 'fas fa-exclamation-circle';
            case 'warning': return 'fas fa-exclamation-triangle';
            case 'info': return 'fas fa-info-circle';
            default: return 'fas fa-info-circle';
        }
    }

    remove(id) {
        const toastIndex = this.toasts.findIndex(t => t.id === id);
        if (toastIndex === -1) return;

        const toast = this.toasts[toastIndex];

        // Prevent multiple removals of the same toast
        if (toast.removing) return;
        toast.removing = true;

        // Clear timeout if exists
        if (toast.timeout) {
            clearTimeout(toast.timeout);
            toast.timeout = null;
        }

        // Add removing class for animation
        if (toast.element && toast.element.classList) {
            toast.element.classList.add('removing');
        }

        // Remove from array immediately to prevent race conditions
        this.toasts.splice(toastIndex, 1);

        // Remove DOM element after animation
        setTimeout(() => {
            if (toast.element && toast.element.parentNode) {
                try {
                    toast.element.parentNode.removeChild(toast.element);
                } catch (error) {
                    console.warn('Error removing toast element:', error);
                }
            }
        }, 300);
    }

    // Clear all toasts
    clearAll() {
        // Clear all timeouts
        this.toasts.forEach(toast => {
            if (toast.timeout) {
                clearTimeout(toast.timeout);
            }
        });

        // Remove all DOM elements
        this.toasts.forEach(toast => {
            if (toast.element && toast.element.parentNode) {
                try {
                    toast.element.parentNode.removeChild(toast.element);
                } catch (error) {
                    console.warn('Error removing toast element:', error);
                }
            }
        });

        // Clear the array
        this.toasts = [];
    }

    // Get count of active toasts
    getCount() {
        return this.toasts.length;
    }
}

// Initialize toast manager
window.toastManager = new ToastManager();
window.showToast = (message, type, duration) => window.toastManager.show(message, type, duration);

// Add cleanup function for toasts
window.clearAllToasts = () => window.toastManager.clearAll();

// Date Utilities
window.dateUtils = {
    formatRelativeTime(dateInput) {
        const date = dateInput instanceof Date ? dateInput : new Date(dateInput);
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);

        if (isNaN(date.getTime())) {
            return 'Invalid date';
        }

        if (diffInSeconds < 60) {
            return 'Just now';
        }

        const diffInMinutes = Math.floor(diffInSeconds / 60);
        if (diffInMinutes < 60) {
            return `${diffInMinutes}m`;
        }

        const diffInHours = Math.floor(diffInMinutes / 60);
        if (diffInHours < 24) {
            return `${diffInHours}h`;
        }

        const diffInDays = Math.floor(diffInHours / 24);
        if (diffInDays < 7) {
            return `${diffInDays}d`;
        }

        return date.toLocaleDateString();
    }
};

// Emoji Utilities
window.emojiUtils = {
    POPULAR_EMOJIS: [
        "😀", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "😊", "😇",
        "🙂", "🙃", "😉", "😌", "😍", "🥰", "😘", "😗", "😙", "😚",
        "😋", "😛", "😝", "😜", "🤪", "🤨", "🧐", "🤓", "😎", "🤩",
        "😏", "😒", "😞", "😔", "😟", "😕", "🙁", "☹️", "😣", "😖",
        "😫", "😩", "🥺", "😢", "😭", "😤", "😠", "😡", "🤬", "🤯",
        "❤️", "🧡", "💛", "💚", "💙", "💜", "🖤", "💔", "❣️", "💕",
        "👍", "👎", "👏", "🙌", "👐", "🤲", "🤝", "👊", "✌️", "🤞"
    ],

    EMOJI_COLLECTIONS: {
        smileys: [
            "😀", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "🥲", "☺️",
            "😊", "😇", "🙂", "🙃", "😉", "😌", "😍", "🥰", "😘", "😗",
            "😙", "😚", "😋", "😛", "😝", "😜", "🤪", "🤨", "🧐", "🤓",
            "😎", "🥸", "🤩", "🥳", "😏", "😒", "😞", "😔", "😟", "😕",
            "🙁", "☹️", "😣", "😖", "😫", "😩", "🥺", "😢", "😭", "😤"
        ],
        hearts: [
            "❤️", "🧡", "💛", "💚", "💙", "💜", "🖤", "🤍", "🤎", "💔",
            "❣️", "💕", "💞", "💓", "💗", "💖", "💘", "💝", "💟"
        ],
        handGestures: [
            "👍", "👎", "👊", "✊", "🤛", "🤜", "🤞", "✌️", "🤟", "🤘",
            "👌", "🤌", "🤏", "👈", "👉", "👆", "👇", "☝️", "✋", "🤚",
            "🖐", "🖖", "👋", "🤙", "💪", "🦾", "🖕", "✍️", "🙏", "🦶"
        ]
    },

    getEmojiCollection(category = 'popular') {
        if (category === 'popular') return this.POPULAR_EMOJIS;
        return this.EMOJI_COLLECTIONS[category] || this.POPULAR_EMOJIS;
    }
};

// Message Item Component
window.messageComponents = {
    createMessageElement(message, isOwnMessage, avatar, currentUser) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message-item ${isOwnMessage ? 'sent' : 'received'}`;
        messageDiv.dataset.messageId = message.id;

        const baseUrl = 'http://localhost:8080'; // Your Go backend URL
        const avatarSrc = avatar ? `${baseUrl}/uploads/${avatar}` : './avatar.png';

        messageDiv.innerHTML = `
            ${!isOwnMessage ? `
                <img src="${avatarSrc}" 
                     alt="${message.sender?.firstName || 'User'}" 
                     class="message-avatar">
            ` : ''}
            <div class="message-content">
                <div class="message-text">${this.escapeHtml(message.content)}</div>
                <div class="message-time">
                    ${window.dateUtils.formatRelativeTime(new Date(message.createdAt))}
                    ${isOwnMessage ? `
                        <span class="read-status">
                            ${message.isRead ?
                    '<i class="fas fa-check-double" title="Read"></i>' :
                    '<i class="fas fa-check" title="Sent"></i>'
                }
                        </span>
                    ` : ''}
                </div>
            </div>
        `;

        return messageDiv;
    },

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};

// Conversation Item Component
window.conversationComponents = {
    createConversationElement(contact, isSelected, currentUser) {
        const baseUrl = 'http://localhost:8080';
        const avatarSrc = contact.avatar ? `${baseUrl}/uploads/${contact.avatar}` : './avatar.png';

        const conversationDiv = document.createElement('div');
        conversationDiv.className = `conversation-item ${isSelected ? 'selected' : ''} ${contact.unreadCount > 0 ? 'unread' : ''}`;
        conversationDiv.dataset.userId = contact.userId;

        const lastMessagePreview = this.getLastMessagePreview(contact, currentUser);
        const timeDisplay = contact.lastSent && contact.lastSent !== '0001-01-01T00:00:00Z'
            ? window.dateUtils.formatRelativeTime(new Date(contact.lastSent))
            : '';

        conversationDiv.innerHTML = `
            <div class="avatar-container">
                <img src="${avatarSrc}" 
                     alt="${contact.firstName} ${contact.lastName}" 
                     class="avatar">
                <span class="status-indicator ${window.userStatusService ? window.userStatusService.isUserOnline(contact.userId, contact.isOnline) ? 'online' : 'offline' : contact.isOnline ? 'online' : 'offline'}"></span>
            </div>
            <div class="preview-content">
                <h4 class="${contact.unreadCount > 0 ? 'bold-name' : ''}">
                    ${contact.firstName} ${contact.lastName}
                </h4>
                <p class="${contact.unreadCount > 0 ? 'bold-preview' : 'message-preview'}">
                    ${lastMessagePreview}
                </p>
            </div>
            <div class="conversation-meta">
                <span class="timestamp ${contact.unreadCount > 0 ? 'bold-timestamp' : ''}">
                    ${timeDisplay}
                </span>
                ${contact.unreadCount > 0 ? `
                    <span class="unread-badge">${contact.unreadCount}</span>
                ` : ''}
            </div>
        `;

        return conversationDiv;
    },

    getLastMessagePreview(contact, currentUser) {
        if (!contact.lastMessage) {
            return 'Start a conversation';
        }

        if (contact.lastMessageSenderId === currentUser?.id) {
            const preview = contact.lastMessage.length > 30
                ? `${contact.lastMessage.substring(0, 30)}...`
                : contact.lastMessage;
            return `You: ${preview}`;
        }

        return contact.lastMessage.length > 30
            ? `${contact.lastMessage.substring(0, 30)}...`
            : contact.lastMessage;
    }
};

// Contact Item Component (for new message modal)
window.contactComponents = {
    createContactElement(contact) {
        const baseUrl = 'http://localhost:8080';
        const avatarSrc = contact.image || './avatar.png';

        const contactDiv = document.createElement('div');
        contactDiv.className = 'contact-item';
        contactDiv.dataset.contactId = contact.contactId || contact.id;

        contactDiv.innerHTML = `
            <div class="contact-avatar">
                <img src="${avatarSrc}" alt="${contact.name}" class="avatar">
                <span class="status-dot ${contact.isOnline ? 'online' : 'offline'}"></span>
            </div>
            <div class="contact-info">
                <h4>${contact.name}</h4>
                ${contact.mutualFriends ? `<p>${contact.mutualFriends} mutual friends</p>` : ''}
            </div>
        `;

        return contactDiv;
    }
};

// Loading Spinner Component
window.spinnerComponents = {
    createSpinner(size = 'medium', color = 'primary') {
        const spinner = document.createElement('div');
        spinner.className = `spinner ${size} ${color}`;

        spinner.innerHTML = `
            <div class="bounce1"></div>
            <div class="bounce2"></div>
            <div class="bounce3"></div>
        `;

        return spinner;
    },

    createFullPageSpinner(message = 'Loading...') {
        const overlay = document.createElement('div');
        overlay.className = 'full-page-overlay';

        overlay.innerHTML = `
            <div class="loading-content">
                <div class="spinner large primary">
                    <div class="bounce1"></div>
                    <div class="bounce2"></div>
                    <div class="bounce3"></div>
                </div>
                <p>${message}</p>
            </div>
        `;

        return overlay;
    }
};

// Search Utilities
window.searchUtils = {
    filterContacts(contacts, searchQuery) {
        if (!searchQuery.trim()) return contacts;

        const query = searchQuery.toLowerCase();
        return contacts.filter(contact => {
            if (!contact || !contact.userId) return false;

            const firstName = (contact.firstName || '').toLowerCase();
            const lastName = (contact.lastName || '').toLowerCase();
            const name = (contact.name || '').toLowerCase();

            return firstName.includes(query) ||
                lastName.includes(query) ||
                name.includes(query);
        });
    },

    filterMessages(messages, searchQuery) {
        if (!searchQuery.trim()) return messages;

        const query = searchQuery.toLowerCase();
        return messages.filter(message =>
            message.content && message.content.toLowerCase().includes(query)
        );
    },

    highlightText(text, searchQuery) {
        if (!searchQuery.trim()) return text;

        const regex = new RegExp(`(${searchQuery})`, 'gi');
        return text.replace(regex, '<mark>$1</mark>');
    }
};

// Form Utilities
window.formUtils = {
    validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    },

    validatePassword(password) {
        return password && password.length >= 6;
    },

    showFieldError(fieldId, message) {
        const field = document.getElementById(fieldId);
        if (!field) return;

        // Remove existing error
        this.clearFieldError(fieldId);

        // Add error class
        field.classList.add('error');

        // Create error message
        const errorDiv = document.createElement('div');
        errorDiv.className = 'field-error';
        errorDiv.textContent = message;
        errorDiv.id = `${fieldId}-error`;

        // Insert after field
        field.parentNode.insertBefore(errorDiv, field.nextSibling);
    },

    clearFieldError(fieldId) {
        const field = document.getElementById(fieldId);
        const errorElement = document.getElementById(`${fieldId}-error`);

        if (field) {
            field.classList.remove('error');
        }

        if (errorElement) {
            errorElement.remove();
        }
    },

    clearAllErrors() {
        const errorElements = document.querySelectorAll('.field-error');
        errorElements.forEach(el => el.remove());

        const errorFields = document.querySelectorAll('.error');
        errorFields.forEach(field => field.classList.remove('error'));
    }
};

// Network Status Utilities
window.networkUtils = {
    isOnline() {
        return navigator.onLine;
    },

    onStatusChange(callback) {
        window.addEventListener('online', () => callback(true));
        window.addEventListener('offline', () => callback(false));
    },

    checkConnectivity() {
        return new Promise((resolve) => {
            if (!navigator.onLine) {
                resolve(false);
                return;
            }

            // Try to fetch a small resource to verify connectivity
            fetch('/favicon.ico', {
                method: 'HEAD',
                cache: 'no-cache'
            })
                .then(() => resolve(true))
                .catch(() => resolve(false));
        });
    }
};

// Emoji Picker Utilities
window.emojiPicker = {
    currentCategory: 'popular',

    populateEmojiGrid(category = 'popular') {
        const grid = document.getElementById('emoji-grid');
        if (!grid) return;

        const emojis = window.emojiUtils.getEmojiCollection(category);
        grid.innerHTML = '';

        emojis.forEach(emoji => {
            const button = document.createElement('button');
            button.type = 'button';
            button.className = 'emoji-button';
            button.textContent = emoji;
            button.onclick = () => this.selectEmoji(emoji);
            grid.appendChild(button);
        });

        this.currentCategory = category;
    },

    selectEmoji(emoji) {
        const messageInput = document.getElementById('message-input');
        if (messageInput) {
            messageInput.value += emoji;
            messageInput.focus();
        }

        this.hide();
    },

    show() {
        const picker = document.getElementById('emoji-picker');
        if (picker) {
            picker.style.display = 'block';
            this.populateEmojiGrid(this.currentCategory);
        }
    },

    hide() {
        const picker = document.getElementById('emoji-picker');
        if (picker) {
            picker.style.display = 'none';
        }
    },

    toggle() {
        const picker = document.getElementById('emoji-picker');
        if (picker) {
            if (picker.style.display === 'none' || !picker.style.display) {
                this.show();
            } else {
                this.hide();
            }
        }
    }
};

// Initialize utilities when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('Utilities initialized');

    // Set up network status monitoring
    window.networkUtils.onStatusChange((isOnline) => {
        console.log('Network status changed:', isOnline ? 'online' : 'offline');

        if (isOnline) {
            window.showToast('Connection restored', 'success');
        } else {
            window.showToast('Connection lost - you are now offline', 'warning');
        }
    });
});