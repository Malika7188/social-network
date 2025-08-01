// Router Service - Simple routing for the Electron app

class Router {
    constructor() {
        this.currentRoute = 'loading';
        this.listeners = new Set();

        // Bind methods
        this.navigate = this.navigate.bind(this);
        this.init = this.init.bind(this);
    }

    init() {
        // Initial route determination
        this.determineInitialRoute();
    }

    async determineInitialRoute() {
        // Show loading screen
        this.showSection('loading');

        // Wait for auth service to initialize with retries
        let attempts = 0;
        const maxAttempts = 20;

        while (attempts < maxAttempts) {
            if (window.authService && !window.authService.loading) {
                break;
            }
            await new Promise(resolve => setTimeout(resolve, 250));
            attempts++;
        }

        if (attempts >= maxAttempts) {
            console.error('Auth service failed to initialize');
            this.navigate('auth');
            return;
        }

        // Navigate based on authentication status
        if (window.authService && window.authService.isAuthenticated()) {
            this.navigate('messenger');
        } else {
            this.navigate('auth');
        }
    }

    navigate(route) {
        const previousRoute = this.currentRoute;
        this.currentRoute = route;

        this.showSection(route);
        this.notifyListeners(route, previousRoute);

        console.log(`Navigated from ${previousRoute} to ${route}`);
    }

    showSection(route) {
        // Hide all sections first
        this.hideAllSections();

        switch (route) {
            case 'loading':
                this.showElement('loading-screen');
                break;
            case 'auth':
                this.showElement('app-container');
                this.showElement('auth-section');
                this.hideElement('messenger-section');
                break;
            case 'messenger':
                this.showElement('app-container');
                this.showElement('messenger-section');
                this.hideElement('auth-section');

                // Initialize messenger components
                this.initializeMessenger();
                break;
            default:
                console.warn(`Unknown route: ${route}`);
                this.navigate('auth');
        }
    }

    hideAllSections() {
        const sections = [
            'loading-screen',
            'auth-section',
            'messenger-section'
        ];

        sections.forEach(id => this.hideElement(id));
    }

    showElement(id) {
        const element = document.getElementById(id);
        if (element) {
            element.style.display = 'flex';
        }
    }

    hideElement(id) {
        const element = document.getElementById(id);
        if (element) {
            element.style.display = 'none';
        }
    }

    async initializeMessenger() {
        // Initialize WebSocket connection
        if (window.websocketService && !window.websocketService.isConnected()) {
            window.websocketService.connect();
        }

        // Initialize chat service
        if (window.chatService && !window.chatService.isInitialized) {
            window.chatService.init();
        }

        // Load initial data
        if (window.messengerUI) {
            await window.messengerUI.loadInitialData();
        }
    }

    getCurrentRoute() {
        return this.currentRoute;
    }

    // Event listener management
    addListener(callback) {
        this.listeners.add(callback);
        return () => this.listeners.delete(callback);
    }

    notifyListeners(newRoute, previousRoute) {
        this.listeners.forEach(callback => {
            try {
                callback({ newRoute, previousRoute });
            } catch (error) {
                console.error('Router listener error:', error);
            }
        });
    }
}

// Create global instance
window.router = new Router();