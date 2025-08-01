// Initialization Script - Handles proper startup sequence

class AppInitializer {
    constructor() {
        this.initialized = false;
        this.initPromise = null;
    }

    async init() {
        if (this.initPromise) {
            return this.initPromise;
        }

        this.initPromise = this._initialize();
        return this.initPromise;
    }

    async _initialize() {
        try {
            console.log('Starting app initialization...');

            // Step 1: Wait for DOM to be fully ready
            await this.waitForDOM();
            console.log('✓ DOM ready');

            // Step 2: Wait for Electron API to be available
            await this.waitForElectronAPI();
            console.log('✓ Electron API ready');

            // Step 3: Initialize services in correct order
            await this.initializeServices();
            console.log('✓ Services initialized');

            // Step 4: Initialize main app
            await this.initializeApp();
            console.log('✓ App initialized');

            this.initialized = true;
            console.log('🎉 App initialization complete!');

        } catch (error) {
            console.error('❌ App initialization failed:', error);
            this.showCriticalError(error);
            throw error;
        }
    }

    async waitForDOM() {
        return new Promise((resolve) => {
            if (document.readyState === 'complete' || document.readyState === 'interactive') {
                resolve();
            } else {
                document.addEventListener('DOMContentLoaded', resolve, { once: true });
            }
        });
    }

    async waitForElectronAPI() {
        let attempts = 0;
        const maxAttempts = 50;

        while (attempts < maxAttempts) {
            if (window.electronAPI) {
                return;
            }
            await new Promise(resolve => setTimeout(resolve, 100));
            attempts++;
        }

        throw new Error('Electron API not available');
    }

    async initializeServices() {
        // Services should already be loaded via script tags
        // Just verify they exist and are properly initialized

        const requiredServices = [
            'authService',
            'websocketService',
            'userStatusService',
            'chatService',
            'router'
        ];

        for (const serviceName of requiredServices) {
            let attempts = 0;
            const maxAttempts = 20;

            while (attempts < maxAttempts) {
                if (window[serviceName]) {
                    console.log(`✓ ${serviceName} loaded`);
                    break;
                }
                await new Promise(resolve => setTimeout(resolve, 100));
                attempts++;
            }

            if (attempts >= maxAttempts) {
                throw new Error(`Service ${serviceName} failed to load`);
            }
        }

        // Initialize user status service
        if (window.userStatusService && typeof window.userStatusService.init === 'function') {
            window.userStatusService.init();
            console.log('✓ User status service initialized');
        }

        // Wait a bit more for services to fully initialize
        await new Promise(resolve => setTimeout(resolve, 200));
    }

    async initializeApp() {
        // Initialize messenger app
        if (!window.messengerApp) {
            window.messengerApp = new MessengerApp();
            window.messengerUI = window.messengerApp; // Alias for router
        }

        await window.messengerApp.init();
    }

    showCriticalError(error) {
        // Remove loading screen
        const loadingScreen = document.getElementById('loading-screen');
        if (loadingScreen) {
            loadingScreen.style.display = 'none';
        }

        // Show error screen
        const errorScreen = document.createElement('div');
        errorScreen.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 10000;
            color: white;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        `;

        errorScreen.innerHTML = `
            <div style="text-align: center; max-width: 400px; padding: 2rem;">
                <i class="fas fa-exclamation-triangle" style="font-size: 4rem; margin-bottom: 1rem; color: #ff6b6b;"></i>
                <h1 style="margin-bottom: 1rem;">Initialization Failed</h1>
                <p style="margin-bottom: 2rem; opacity: 0.9;">
                    The application failed to start properly. Please try restarting the app.
                </p>
                <button onclick="location.reload()" style="
                    background: #ff6b6b;
                    color: white;
                    border: none;
                    padding: 0.75rem 1.5rem;
                    border-radius: 0.5rem;
                    cursor: pointer;
                    font-size: 1rem;
                ">
                    Restart App
                </button>
                <details style="margin-top: 2rem; text-align: left;">
                    <summary style="cursor: pointer; margin-bottom: 0.5rem;">Technical Details</summary>
                    <pre style="background: rgba(0,0,0,0.2); padding: 1rem; border-radius: 0.25rem; font-size: 0.8rem; overflow: auto;">
${error.message}
${error.stack || ''}
                    </pre>
                </details>
            </div>
        `;

        document.body.appendChild(errorScreen);
    }
}

// Global initializer instance
window.appInitializer = new AppInitializer();

// Start initialization when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        window.appInitializer.init().catch(console.error);
    });
} else {
    // DOM already ready
    window.appInitializer.init().catch(console.error);
}