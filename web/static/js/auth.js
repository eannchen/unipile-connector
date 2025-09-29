// Authentication JavaScript
let currentUser = null;
let authToken = null;

// Check if user is logged in
function checkAuth() {
    const token = localStorage.getItem('authToken');
    const userId = localStorage.getItem('userId');
    const username = localStorage.getItem('username');

    if (token && userId && username) {
        authToken = token;
        currentUser = { id: userId, username: username };
        return true;
    }
    return false;
}

// Get auth headers for API requests
function getAuthHeaders() {
    const headers = {
        'Content-Type': 'application/json'
    };

    if (authToken) {
        headers['Authorization'] = `Bearer ${authToken}`;
    }

    return headers;
}

// Show alert messages
function showAlert(message, type = 'info') {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
    alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;

    const container = document.querySelector('.container');
    container.insertBefore(alertDiv, container.firstChild);

    // Auto-dismiss after 5 seconds
    setTimeout(() => {
        alertDiv.remove();
    }, 5000);
}

// Handle login form submission
document.addEventListener('DOMContentLoaded', function () {
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', async function (e) {
            e.preventDefault();

            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;

            try {
                const response = await fetch('/api/v1/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ username, password })
                });

                const data = await response.json();

                if (response.ok) {
                    // Store token and user info
                    authToken = data.token;
                    localStorage.setItem('authToken', data.token);
                    localStorage.setItem('userId', data.user.id);
                    localStorage.setItem('username', data.user.username);

                    showAlert('Login successful! Redirecting to dashboard...', 'success');

                    // Immediate redirect to dashboard
                    setTimeout(() => {
                        window.location.replace('/dashboard');
                    }, 500);
                } else {
                    showAlert(data.error || 'Login failed', 'danger');
                }
            } catch (error) {
                showAlert('Network error. Please try again.', 'danger');
            }
        });
    }

    // Handle register form submission
    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', async function (e) {
            e.preventDefault();

            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const confirmPassword = document.getElementById('confirmPassword').value;

            if (password !== confirmPassword) {
                showAlert('Passwords do not match', 'danger');
                return;
            }

            try {
                const response = await fetch('/api/v1/auth/register', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ username, password })
                });

                const data = await response.json();

                if (response.ok) {
                    showAlert('Registration successful! Redirecting to login...', 'success');
                    setTimeout(() => {
                        window.location.replace('/login');
                    }, 1000);
                } else {
                    showAlert(data.error || 'Registration failed', 'danger');
                }
            } catch (error) {
                showAlert('Network error. Please try again.', 'danger');
            }
        });
    }
});

// Global authentication check - runs on every page
function globalAuthCheck() {
    // Only run on non-dashboard pages
    if (window.location.pathname !== '/dashboard') {
        if (checkAuth()) {
            console.log('Global auth check: User is logged in, redirecting to dashboard');
            window.location.replace('/dashboard');
        }
    } else {
        // On dashboard page, ensure user is authenticated
        if (!checkAuth()) {
            console.log('Global auth check: User not authenticated on dashboard, redirecting to login');
            window.location.replace('/login');
        }
    }
}

// Initialize global auth check
document.addEventListener('DOMContentLoaded', function () {
    // Run immediately
    globalAuthCheck();

    // Also run after a short delay to catch any race conditions
    setTimeout(globalAuthCheck, 100);
});

// Check if user is logged in and redirect if needed
function checkAuthAndRedirect() {
    if (checkAuth()) {
        // User is logged in, redirect to dashboard
        console.log('User is already logged in, redirecting to dashboard');
        window.location.href = '/dashboard';
    }
}

// Force redirect to dashboard for logged-in users
function redirectToDashboardIfLoggedIn() {
    if (checkAuth()) {
        console.log('Redirecting logged-in user to dashboard');
        window.location.replace('/dashboard');
    }
}

// Check authentication on page load and redirect
function initAuthCheck() {
    // Small delay to ensure DOM is ready
    setTimeout(() => {
        redirectToDashboardIfLoggedIn();
    }, 100);
}

// Enhanced logout function
async function logout() {
    if (!confirm('Are you sure you want to logout?')) {
        return;
    }

    try {
        // Call logout endpoint (optional, mainly for server-side logging)
        if (authToken) {
            await fetch('/api/v1/auth/logout', {
                method: 'POST',
                headers: getAuthHeaders()
            });
        }
    } catch (error) {
        console.warn('Logout request failed:', error);
    } finally {
        // Clear local storage regardless of server response
        authToken = null;
        localStorage.removeItem('authToken');
        localStorage.removeItem('userId');
        localStorage.removeItem('username');
        currentUser = null;

        // Redirect to home page
        window.location.href = '/';
    }
}
