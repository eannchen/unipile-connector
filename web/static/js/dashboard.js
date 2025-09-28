// Dashboard JavaScript
let currentUser = null;
let currentAccountID = null;
let authToken = null;

// Check authentication and load dashboard
document.addEventListener('DOMContentLoaded', function () {
    const token = localStorage.getItem('authToken');
    const userId = localStorage.getItem('userId');
    const username = localStorage.getItem('username');

    if (!token || !userId || !username) {
        console.log('User not authenticated, redirecting to login');
        window.location.replace('/login');
        return;
    }

    authToken = token;
    currentUser = { id: userId, username: username };

    // Update user info in navbar
    document.getElementById('userInfo').textContent = `Welcome, ${username}!`;

    // Load user accounts
    loadUserAccounts();

    // Setup connection type toggle
    setupConnectionTypeToggle();
});

// Handle navbar brand click - stay on dashboard when logged in
function handleNavbarBrandClick(event) {
    event.preventDefault();
    // When logged in, clicking the brand should stay on dashboard
    // Just scroll to top or refresh the dashboard
    window.scrollTo({ top: 0, behavior: 'smooth' });
}

// Enhanced logout function
async function logout() {
    if (!confirm('Are you sure you want to logout?')) {
        return;
    }

    try {
        // Show loading state
        const logoutBtn = document.querySelector('button[onclick="logout()"]');
        const originalText = logoutBtn.textContent;
        logoutBtn.textContent = 'Logging out...';
        logoutBtn.disabled = true;

        // Call logout endpoint
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
        localStorage.removeItem('userEmail');
        currentUser = null;

        // Show success message briefly before redirect
        showAlert('Logged out successfully!', 'success');

        // Redirect to home page after a short delay
        setTimeout(() => {
            window.location.href = '/';
        }, 1000);
    }
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

// Setup connection type toggle
function setupConnectionTypeToggle() {
    const credentialsRadio = document.getElementById('credentials');
    const cookieRadio = document.getElementById('cookie');
    const credentialsForm = document.getElementById('credentialsForm');
    const cookieForm = document.getElementById('cookieForm');

    credentialsRadio.addEventListener('change', function () {
        credentialsForm.style.display = 'block';
        cookieForm.style.display = 'none';
    });

    cookieRadio.addEventListener('change', function () {
        credentialsForm.style.display = 'none';
        cookieForm.style.display = 'block';
    });
}

// Load user accounts
async function loadUserAccounts() {
    try {
        const response = await fetch('/api/v1/accounts', {
            method: 'GET',
            headers: getAuthHeaders()
        });

        if (response.ok) {
            const data = await response.json();
            displayAccounts(data.accounts);
        } else if (response.status === 401) {
            // Token expired, redirect to login
            window.location.href = '/login';
        } else {
            console.error('Failed to load accounts');
        }
    } catch (error) {
        console.error('Error loading accounts:', error);
    }
}

// Display accounts
function displayAccounts(accounts) {
    const accountsList = document.getElementById('accountsList');
    const disconnectBtn = document.getElementById('disconnectBtn');

    if (accounts && accounts.length > 0) {
        let html = '<div class="list-group">';
        accounts.forEach(account => {
            html += `
                <div class="list-group-item">
                    <div class="d-flex w-100 justify-content-between">
                        <h5 class="mb-1">${account.provider.charAt(0).toUpperCase() + account.provider.slice(1)}</h5>
                        <small>${new Date(account.created_at).toLocaleDateString()}</small>
                    </div>
                    <p class="mb-1">Account ID: ${account.account_id}</p>
                    <button type="button" class="btn btn-outline-danger btn-sm" onclick="disconnectLinkedIn()"
                         id="disconnectBtn">
                        Disconnect LinkedIn
                    </button>
                </div>
            `;
        });
        html += '</div>';
        accountsList.innerHTML = html;

        // Show disconnect button if LinkedIn account exists
        const linkedinAccount = accounts.find(acc => acc.provider === 'linkedin');
        if (linkedinAccount) {
            disconnectBtn.style.display = 'block';
        }
    } else {
        accountsList.innerHTML = '<p class="text-muted">No accounts connected yet.</p>';
        disconnectBtn.style.display = 'none';
    }
}

// Connect LinkedIn account
async function connectLinkedIn() {
    const connectionType = document.querySelector('input[name="connectionType"]:checked').value;
    let requestData = { type: connectionType };

    if (connectionType === 'credentials') {
        const username = document.getElementById('linkedinUsername').value;
        const password = document.getElementById('linkedinPassword').value;

        if (!username || !password) {
            showAlert('Please enter LinkedIn username and password', 'danger');
            return;
        }

        requestData.username = username;
        requestData.password = password;
    } else if (connectionType === 'cookie') {
        const accessToken = document.getElementById('linkedinAccessToken').value;
        const userAgent = document.getElementById('userAgent').value;

        if (!accessToken) {
            showAlert('Please enter LinkedIn access token', 'danger');
            return;
        }

        requestData.access_token = accessToken;
        if (userAgent) {
            requestData.user_agent = userAgent;
        }
    }

    try {
        const response = await fetch('/api/v1/accounts/linkedin/connect', {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify(requestData)
        });

        const data = await response.json();

        if (response.status === 202) {
            // Checkpoint required
            currentAccountID = data.account_id;
            showCheckpointSection(data.checkpoint);
            showAlert('Checkpoint required: ' + data.checkpoint.type, 'info');
        } else if (response.ok) {
            showAlert('LinkedIn account connected successfully!', 'success');
            // Clear form
            document.getElementById('linkedinUsername').value = '';
            document.getElementById('linkedinPassword').value = '';
            document.getElementById('linkedinAccessToken').value = '';
            document.getElementById('userAgent').value = '';
            // Reload accounts
            loadUserAccounts();
        } else {
            showAlert(data.error || 'Failed to connect LinkedIn account', 'danger');
        }
    } catch (error) {
        showAlert('Network error. Please try again.', 'danger');
    }
}

// Show checkpoint section
function showCheckpointSection(checkpoint) {
    const checkpointSection = document.getElementById('checkpointSection');
    const checkpointType = document.getElementById('checkpointType');

    checkpointType.textContent = checkpoint.type;
    checkpointSection.style.display = 'block';

    // Scroll to checkpoint section
    checkpointSection.scrollIntoView({ behavior: 'smooth' });
}

// Solve checkpoint
async function solveCheckpoint() {
    const code = document.getElementById('checkpointCode').value;

    if (!code) {
        showAlert('Please enter verification code', 'danger');
        return;
    }

    try {
        const response = await fetch('/api/v1/accounts/linkedin/checkpoint', {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({
                account_id: currentAccountID,
                code: code
            })
        });

        const data = await response.json();

        if (response.status === 202) {
            // Another checkpoint required
            showAlert('Another checkpoint required: ' + data.checkpoint.type, 'info');
            document.getElementById('checkpointCode').value = '';
        } else if (response.ok) {
            showAlert('LinkedIn account connected successfully!', 'success');
            hideCheckpointSection();
            loadUserAccounts();
        } else {
            showAlert(data.error || 'Failed to solve checkpoint', 'danger');
        }
    } catch (error) {
        showAlert('Network error. Please try again.', 'danger');
    }
}

// Cancel checkpoint
function cancelCheckpoint() {
    hideCheckpointSection();
    currentAccountID = null;
}

// Hide checkpoint section
function hideCheckpointSection() {
    const checkpointSection = document.getElementById('checkpointSection');
    checkpointSection.style.display = 'none';
    document.getElementById('checkpointCode').value = '';
}

// Disconnect LinkedIn account
async function disconnectLinkedIn() {
    if (!confirm('Are you sure you want to disconnect your LinkedIn account?')) {
        return;
    }

    try {
        const response = await fetch('/api/v1/accounts/linkedin', {
            method: 'DELETE',
            headers: getAuthHeaders()
        });

        const data = await response.json();

        if (response.ok) {
            showAlert('LinkedIn account disconnected successfully!', 'success');
            loadUserAccounts();
        } else {
            showAlert(data.error || 'Failed to disconnect LinkedIn account', 'danger');
        }
    } catch (error) {
        showAlert('Network error. Please try again.', 'danger');
    }
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
