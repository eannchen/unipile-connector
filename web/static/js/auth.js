// Authentication JavaScript
let currentUser = null;

// Check if user is logged in
function checkAuth() {
    const userId = localStorage.getItem('userId');
    const username = localStorage.getItem('username');

    if (userId && username) {
        currentUser = { id: userId, username: username };
        return true;
    }
    return false;
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
                    localStorage.setItem('userId', data.user.id);
                    localStorage.setItem('username', data.user.username);
                    showAlert('Login successful!', 'success');
                    setTimeout(() => {
                        window.location.href = '/dashboard';
                    }, 1000);
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
            const email = document.getElementById('email').value;
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
                    body: JSON.stringify({ username, email, password })
                });

                const data = await response.json();

                if (response.ok) {
                    showAlert('Registration successful! Please login.', 'success');
                    setTimeout(() => {
                        window.location.href = '/login';
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

// Logout function
function logout() {
    localStorage.removeItem('userId');
    localStorage.removeItem('username');
    currentUser = null;
    window.location.href = '/';
}

