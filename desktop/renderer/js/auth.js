// Authentication Service - Vanilla JS version of authcontext.js

class AuthService {
    constructor() {
        this.currentUser = null;
        this.token = null;
        this.loading = false;
        this.listeners = new Set();
        this.API_URL = 'http://localhost:8080/api';

        // Initialize from stored data
        this.init();
    }

    async init() {
        this.loading = true;
        this.notifyListeners();

        try {
            // Get stored data from Electron store
            const userData = await window.electronAPI.store.get('userData');
            const token = await window.electronAPI.store.get('token');

            if (userData && token) {
                this.currentUser = userData;
                this.token = token;

                // Validate token
                await this.validateToken(token);
            } else {
                this.loading = false;
                this.notifyListeners();
            }
        } catch (error) {
            console.error('Auth initialization error:', error);
            this.loading = false;
            this.notifyListeners();
        }
    }

    async validateToken(token) {
        try {
            const response = await fetch(`${this.API_URL}/auth/validate_token`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                    'X-Client-Type': 'electron'
                }
            });

            if (!response.ok) {
                await this.logout(false);
                return false;
            }

            this.loading = false;
            this.notifyListeners();
            return true;
        } catch (error) {
            console.error('Token validation error:', error);
            await this.logout(false);
            return false;
        }
    }

    async login(formData) {
        try {
            this.loading = true;
            this.notifyListeners();

            const response = await fetch(`${this.API_URL}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Client-Type': 'electron'
                },
                body: JSON.stringify({
                    email: formData.email,
                    password: formData.password
                })
            });

            const data = await response.json();

            if (data && data.user && data.token) {
                // Store in Electron store
                await window.electronAPI.store.set('userData', data.user);
                await window.electronAPI.store.set('token', data.token);

                this.currentUser = data.user;
                this.token = data.token;
                this.loading = false;

                this.notifyListeners();
                window.showToast('Logged in successfully!', 'success');
                return true;
            } else {
                const errorMessage = data.message || data.error || 'Login failed';
                window.showToast(errorMessage, 'error');
                this.loading = false;
                this.notifyListeners();
                return false;
            }
        } catch (error) {
            this.loading = false;
            this.notifyListeners();
            console.error('Login error:', error);
            window.showToast('Login failed. Please try again.', 'error');
            return false;
        }
    }

    async logout(sendRequest = true) {
        try {
            // Send user_away message to WebSocket if connected
            if (window.websocketService && window.websocketService.isConnected()) {
                window.websocketService.sendMessage({ type: 'user_away' });
                await new Promise(resolve => setTimeout(resolve, 100));
            }

            // Close WebSocket connection
            if (window.websocketService) {
                window.websocketService.disconnect();
            }

            if (sendRequest && this.token) {
                try {
                    await fetch(`${this.API_URL}/auth/logout`, {
                        method: 'POST',
                        headers: {
                            'Authorization': `Bearer ${this.token}`,
                            'Content-Type': 'application/json',
                            'X-Client-Type': 'electron'
                        }
                    });
                } catch (error) {
                    console.error('Logout request failed:', error);
                }
            }

            // Clear stored data
            await window.electronAPI.store.delete('userData');
            await window.electronAPI.store.delete('token');

            // Clear all toasts
            if (window.clearAllToasts) {
                window.clearAllToasts();
            }

            // Clear user status notifications
            if (window.userStatusService && window.userStatusService.recentNotifications) {
                window.userStatusService.recentNotifications.clear();
            }

            this.currentUser = null;
            this.token = null;
            this.loading = false;

            this.notifyListeners();

            if (sendRequest && window.router) {
                window.router.navigate('auth');
            }
        } catch (error) {
            console.error('Logout error:', error);
            this.loading = false;
            this.notifyListeners();

            // Force navigation to auth even if there's an error
            if (sendRequest && window.router) {
                window.router.navigate('auth');
            }
        }
    }

    async authenticatedFetch(url, options = {}) {
        try {
            if (!this.token) {
                await this.logout(true);
                throw new Error('No token found');
            }

            const authHeaders = {
                'Authorization': `Bearer ${this.token}`,
                'Content-Type': 'application/json',
                'X-Client-Type': 'electron' // Identify as Electron client
            };

            const response = await fetch(`${this.API_URL}/${url}`, {
                ...options,
                headers: {
                    ...authHeaders,
                    ...options.headers
                }
            });

            if (response.status === 401) {
                await this.logout(true);
                throw new Error('Unauthorized - Please log in again.');
            }

            return response;
        } catch (error) {
            console.error('Authenticated fetch error:', error);
            throw error;
        }
    }

    getAuthHeader() {
        return this.token ? { 'Authorization': `Bearer ${this.token}` } : {};
    }

    isAuthenticated() {
        return !!this.token && !!this.currentUser;
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
                    currentUser: this.currentUser,
                    token: this.token,
                    loading: this.loading,
                    isAuthenticated: this.isAuthenticated()
                });
            } catch (error) {
                console.error('Listener error:', error);
            }
        });
    }

    // Redirect to social network website for registration
    redirectToRegister() {
        const registerUrl = 'http://localhost:3000/register'; // Your Next.js frontend URL

        if (window.electronAPI && window.electronAPI.openExternal) {
            window.electronAPI.openExternal(registerUrl);
        } else {
            // Fallback - try to open in same window (will be blocked but shows intent)
            console.log('Would open registration URL:', registerUrl);
            window.showToast('Please restart the app and try again', 'error');
        }
    }
}

// Create global instance
window.authService = new AuthService();