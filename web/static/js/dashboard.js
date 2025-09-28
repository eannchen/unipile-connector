// Dashboard JavaScript
let currentUser = null;

// Check authentication and load dashboard
document.addEventListener('DOMContentLoaded', function () {
    const userId = localStorage.getItem('userId');
    const username = localStorage.getItem('username');

    if (!userId || !username) {
        window.location.href = '/login';
        return;
    }

    currentUser = { id: userId, username: username };

    // Update user info in navbar
    document.getElementById('userInfo').textContent = `Welcome, ${username}!`;

    // Load user accounts
    loadUserAccounts();

    // Setup connection type toggle
    setupConnectionTypeToggle();
});

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
            headers: {
                'X-User-ID': currentUser.id
            }
        });

        if (response.ok) {
            const data = await response.json();
            displayAccounts(data.accounts);
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
        const cookie = document.getElementById('linkedinCookie').value;

        if (!cookie) {
            showAlert('Please enter LinkedIn cookie', 'danger');
            return;
        }

        requestData.cookie = cookie;
    }

    try {
        const response = await fetch('/api/v1/accounts/linkedin/connect', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': currentUser.id
            },
            body: JSON.stringify(requestData)
        });

        const data = await response.json();

        if (response.ok) {
            showAlert('LinkedIn account connected successfully!', 'success');
            // Clear form
            document.getElementById('linkedinUsername').value = '';
            document.getElementById('linkedinPassword').value = '';
            document.getElementById('linkedinCookie').value = '';
            // Reload accounts
            loadUserAccounts();
        } else {
            showAlert(data.error || 'Failed to connect LinkedIn account', 'danger');
        }
    } catch (error) {
        showAlert('Network error. Please try again.', 'danger');
    }
}

// Disconnect LinkedIn account
async function disconnectLinkedIn() {
    if (!confirm('Are you sure you want to disconnect your LinkedIn account?')) {
        return;
    }

    try {
        const response = await fetch('/api/v1/accounts/linkedin', {
            method: 'DELETE',
            headers: {
                'X-User-ID': currentUser.id
            }
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

