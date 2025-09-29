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

    if (!accountsList) {
        console.error('accountsList element not found');
        return;
    }

    if (accounts && accounts.length > 0) {
        // Separate accounts by status
        const connectedAccounts = accounts.filter(acc => acc.current_status === 'OK');
        const connectingAccounts = accounts.filter(acc => acc.current_status !== 'OK');

        let html = '';

        // Display connected accounts (status = "OK")
        if (connectedAccounts.length > 0) {
            html += '<h6 class="text-success mb-3"><i class="fas fa-check-circle"></i> Connected Accounts</h6>';
            html += '<div class="list-group mb-4">';
            connectedAccounts.forEach(account => {
                html += `
                    <div class="list-group-item list-group-item-success">
                        <div class="d-flex w-100 justify-content-between">
                            <h5 class="mb-1">${account.provider.charAt(0).toUpperCase() + account.provider.slice(1)}</h5>
                            <small>${new Date(account.created_at).toLocaleDateString()}</small>
                        </div>
                        <p class="mb-1">Account ID: ${account.account_id}</p>
                        <button type="button" class="btn btn-outline-danger btn-sm" onclick="disconnectLinkedIn('${account.account_id}')"
                             id="disconnectBtn">
                            Disconnect LinkedIn
                        </button>
                    </div>
                `;
            });
            html += '</div>';
        }

        // Display connecting accounts (status != "OK")
        if (connectingAccounts.length > 0) {
            html += '<h6 class="text-warning mb-3"><i class="fas fa-clock"></i> Connecting Accounts</h6>';
            html += '<div class="list-group">';
            connectingAccounts.forEach(account => {
                const statusBadge = getStatusBadge(account.current_status);
                html += `
                    <div class="list-group-item list-group-item-warning">
                        <div class="d-flex w-100 justify-content-between">
                            <h5 class="mb-1">${account.provider.charAt(0).toUpperCase() + account.provider.slice(1)}</h5>
                            <div>
                                ${statusBadge}
                                <small class="ms-2">${new Date(account.created_at).toLocaleDateString()}</small>
                            </div>
                        </div>
                        <p class="mb-1">Account ID: ${account.account_id}</p>
                        ${getConnectingAccountActions(account)}
                    </div>
                `;
            });
            html += '</div>';
        }

        accountsList.innerHTML = html;

        // Show disconnect button if LinkedIn account exists and is connected
        const linkedinAccount = connectedAccounts.find(acc => acc.provider.toLowerCase() === 'linkedin');
        if (linkedinAccount && disconnectBtn) {
            disconnectBtn.style.display = 'block';
        }
    } else {
        accountsList.innerHTML = '<p class="text-muted">No accounts connected yet.</p>';
        if (disconnectBtn) {
            disconnectBtn.style.display = 'none';
        }
    }
}

// Get status badge for connecting accounts
function getStatusBadge(status) {
    const statusMap = {
        'PENDING': '<span class="badge bg-warning">Pending</span>',
        'CONNECTING': '<span class="badge bg-info">Connecting</span>',
        'ERROR': '<span class="badge bg-danger">Error</span>',
        'STOPPED': '<span class="badge bg-secondary">Stopped</span>',
        'CREDENTIALS': '<span class="badge bg-warning">Credentials</span>',
        'SYNC_SUCCESS': '<span class="badge bg-success">Sync Success</span>',
        'RECONNECTED': '<span class="badge bg-success">Reconnected</span>'
    };
    return statusMap[status] || `<span class="badge bg-secondary">${status}</span>`;
}

// Get actions for connecting accounts
function getConnectingAccountActions(account) {
    if (account.current_status === 'PENDING' && account.account_status_histories && account.account_status_histories.length > 0) {
        const latestHistory = account.account_status_histories[account.account_status_histories.length - 1];
        if (latestHistory.checkpoint) {
            return `
                <div class="mt-2">
                    <button type="button" class="btn btn-primary btn-sm" onclick="resumeCheckpoint('${account.account_id}', '${latestHistory.checkpoint}')">
                        <i class="fas fa-play"></i> Resume Checkpoint
                    </button>
                    <button type="button" class="btn btn-outline-danger btn-sm ms-2" onclick="cancelConnection('${account.account_id}')">
                        <i class="fas fa-times"></i> Cancel
                    </button>
                </div>
            `;
        }
    }

    return `
        <div class="mt-2">
            <button type="button" class="btn btn-outline-danger btn-sm" onclick="cancelConnection('${account.account_id}')">
                <i class="fas fa-times"></i> Cancel Connection
            </button>
        </div>
    `;
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

    // Show loading state
    const connectBtn = document.querySelector('button[onclick="connectLinkedIn()"]');
    let originalText = '';
    if (connectBtn) {
        originalText = connectBtn.innerHTML;
        connectBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Connecting...';
        connectBtn.disabled = true;
    }

    try {
        const response = await fetch('/api/v1/accounts/linkedin/connect', {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify(requestData)
        });

        const data = await response.json();

        if (response.ok) {
            // Check if account status is not "OK" - meaning checkpoint is required
            if (data.current_status && data.current_status !== "OK") {
                // Checkpoint required
                currentAccountID = data.account_id;

                // Check if there's checkpoint information in account status histories
                if (data.account_status_histories && data.account_status_histories.length > 0) {
                    const latestHistory = data.account_status_histories[data.account_status_histories.length - 1];
                    showCheckpointSection({
                        type: latestHistory.checkpoint
                    }, latestHistory.checkpoint_expires_at);
                    showAlert('Checkpoint required: ' + latestHistory.checkpoint, 'info');
                } else {
                    // Generic checkpoint message if no specific checkpoint info
                    showCheckpointSection({
                        type: 'UNKNOWN'
                    }, null);
                    showAlert('Additional verification required', 'info');
                }
            } else {
                // Account connected successfully
                showAlert('LinkedIn account connected successfully!', 'success');
                // Clear form
                document.getElementById('linkedinUsername').value = '';
                document.getElementById('linkedinPassword').value = '';
                document.getElementById('linkedinAccessToken').value = '';
                document.getElementById('userAgent').value = '';
                // Reload accounts
                loadUserAccounts();
            }
        } else {
            showAlert(data.error || 'Failed to connect LinkedIn account', 'danger');
        }
    } catch (error) {
        showAlert('Network error. Please try again.', 'danger');
    } finally {
        // Restore button state
        if (connectBtn && originalText) {
            connectBtn.innerHTML = originalText;
            connectBtn.disabled = false;
        }
    }
}

// Show checkpoint section
function showCheckpointSection(checkpoint, expiresAt) {
    const checkpointSection = document.getElementById('checkpointSection');
    const checkpointType = document.getElementById('checkpointType');
    const checkpointAlert = document.getElementById('checkpointAlert');
    const checkpointCodeInput = document.getElementById('checkpointCode');
    const checkpointLabel = document.querySelector('label[for="checkpointCode"]');

    // Update checkpoint type display
    checkpointType.textContent = checkpoint.type;

    // Update UI based on checkpoint type
    switch (checkpoint.type) {
        case '2FA':
        case 'OTP':
            checkpointAlert.innerHTML = `
                <strong>Two-Factor Authentication Required</strong><br>
                Please enter the 6-digit code from your authenticator app or SMS.
            `;
            checkpointLabel.textContent = 'Verification Code';
            checkpointCodeInput.placeholder = 'Enter 6-digit code';
            checkpointCodeInput.maxLength = 6;
            checkpointCodeInput.type = 'text';
            break;
        case 'IN_APP_VALIDATION':
            checkpointAlert.innerHTML = `
                <strong>LinkedIn App Verification Required</strong><br>
                Please check your LinkedIn mobile app and confirm the connection.
            `;
            checkpointLabel.textContent = 'Confirmation Code (if any)';
            checkpointCodeInput.placeholder = 'Enter confirmation code if prompted';
            checkpointCodeInput.maxLength = 10;
            checkpointCodeInput.type = 'text';
            break;
        case 'CAPTCHA':
            checkpointAlert.innerHTML = `
                <strong>CAPTCHA Verification Required</strong><br>
                Please solve the CAPTCHA challenge.
            `;
            checkpointLabel.textContent = 'CAPTCHA Solution';
            checkpointCodeInput.placeholder = 'Enter CAPTCHA solution';
            checkpointCodeInput.maxLength = 20;
            checkpointCodeInput.type = 'text';
            break;
        case 'PHONE_REGISTER':
            checkpointAlert.innerHTML = `
                <strong>Phone Verification Required</strong><br>
                Please enter your phone number to receive a verification code.
            `;
            checkpointLabel.textContent = 'Phone Number';
            checkpointCodeInput.placeholder = '+1234567890';
            checkpointCodeInput.maxLength = 15;
            checkpointCodeInput.type = 'tel';
            break;
        default:
            checkpointAlert.innerHTML = `
                <strong>Verification Required</strong><br>
                Please complete the verification process.
            `;
            checkpointLabel.textContent = 'Verification Code';
            checkpointCodeInput.placeholder = 'Enter verification code';
            checkpointCodeInput.maxLength = 10;
            checkpointCodeInput.type = 'text';
    }

    // Add expiration timer if provided
    if (expiresAt) {
        const expirationTime = new Date(expiresAt).getTime();
        const now = new Date().getTime();
        const timeLeft = expirationTime - now;

        if (timeLeft > 0) {
            startExpirationTimer(timeLeft);
        }
    }

    checkpointSection.style.display = 'block';
    checkpointCodeInput.value = ''; // Clear previous input

    // Add keyboard support
    checkpointCodeInput.addEventListener('keypress', function (e) {
        if (e.key === 'Enter') {
            solveCheckpoint();
        }
    });

    // Focus on input
    setTimeout(() => {
        checkpointCodeInput.focus();
    }, 300);

    // Scroll to checkpoint section
    checkpointSection.scrollIntoView({ behavior: 'smooth' });
}

// Start expiration timer
function startExpirationTimer(timeLeft) {
    const timerElement = document.createElement('div');
    timerElement.id = 'expirationTimer';
    timerElement.className = 'alert alert-warning mt-2';

    const checkpointAlert = document.getElementById('checkpointAlert');
    checkpointAlert.appendChild(timerElement);

    const timer = setInterval(() => {
        const minutes = Math.floor(timeLeft / 60000);
        const seconds = Math.floor((timeLeft % 60000) / 1000);

        timerElement.innerHTML = `
            <strong>Time remaining:</strong> ${minutes}:${seconds.toString().padStart(2, '0')}
        `;

        timeLeft -= 1000;

        if (timeLeft <= 0) {
            clearInterval(timer);
            timerElement.innerHTML = '<strong>Checkpoint expired!</strong> Please start over.';
            timerElement.className = 'alert alert-danger mt-2';
            document.getElementById('checkpointCode').disabled = true;
            document.querySelector('button[onclick="solveCheckpoint()"]').disabled = true;
        }
    }, 1000);
}

// Validate checkpoint input based on type
function validateCheckpointInput(code, checkpointType) {
    if (!code) {
        return 'Please enter verification code';
    }

    switch (checkpointType) {
        case '2FA':
        case 'OTP':
            if (!/^\d{6}$/.test(code)) {
                return 'Please enter a valid 6-digit code';
            }
            break;
        case 'PHONE_REGISTER':
            if (!/^\+?[\d\s\-\(\)]{10,15}$/.test(code)) {
                return 'Please enter a valid phone number (e.g., +1234567890)';
            }
            break;
        case 'CAPTCHA':
            if (code.length < 3) {
                return 'CAPTCHA solution too short';
            }
            break;
    }

    return null; // No validation error
}

// Solve checkpoint
async function solveCheckpoint() {
    // Check if checkpoint section exists and is visible
    const checkpointSection = document.getElementById('checkpointSection');
    if (!checkpointSection || checkpointSection.style.display === 'none') {
        showAlert('No active checkpoint found. Please try connecting again.', 'danger');
        return;
    }

    const code = document.getElementById('checkpointCode').value;
    const checkpointTypeElement = document.getElementById('checkpointType');
    const checkpointType = checkpointTypeElement ? checkpointTypeElement.textContent : 'UNKNOWN';

    const validationError = validateCheckpointInput(code, checkpointType);
    if (validationError) {
        showAlert(validationError, 'danger');
        return;
    }

    // Check if we have a valid account ID
    if (!currentAccountID) {
        showAlert('No account ID available. Please try connecting again.', 'danger');
        return;
    }

    // Show loading state
    const submitBtn = document.querySelector('button[onclick="solveCheckpoint()"]');
    if (submitBtn) {
        const originalText = submitBtn.innerHTML;
        submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Submitting...';
        submitBtn.disabled = true;
    }

    try {
        const requestBody = {
            account_id: currentAccountID,
            code: code
        };

        const response = await fetch('/api/v1/accounts/linkedin/checkpoint', {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify(requestBody)
        });

        const data = await response.json();

        if (response.ok) {
            showAlert('LinkedIn account connected successfully!', 'success');
            hideCheckpointSection();
            loadUserAccounts();
        } else if (response.status === 401) {
            showAlert('Invalid code or checkpoint expired. Please try again.', 'danger');
            document.getElementById('checkpointCode').value = '';
        } else {
            showAlert(data.error || 'Failed to solve checkpoint', 'danger');
        }
    } catch (error) {
        showAlert('Network error. Please try again.', 'danger');
    } finally {
        // Restore button state
        if (submitBtn) {
            submitBtn.innerHTML = originalText;
            submitBtn.disabled = false;
        }
    }
}

// Test function to manually show checkpoint section
function testCheckpoint() {
    console.log('Testing checkpoint section...');

    const checkpointSection = document.getElementById('checkpointSection');
    const checkpointType = document.getElementById('checkpointType');
    const checkpointCode = document.getElementById('checkpointCode');

    console.log('Checkpoint section:', !!checkpointSection);
    console.log('Checkpoint type element:', !!checkpointType);
    console.log('Checkpoint code element:', !!checkpointCode);

    if (checkpointSection && checkpointType && checkpointCode) {
        // Set test data
        currentAccountID = 'test-account-id-123';
        checkpointType.textContent = '2FA';
        checkpointSection.style.display = 'block';
        checkpointCode.value = '123456';

        console.log('Test checkpoint section shown');
        showAlert('Test checkpoint section activated', 'info');
    } else {
        console.error('Missing checkpoint elements');
        showAlert('Checkpoint elements not found in HTML', 'danger');
    }
}

// Make test function available globally
window.testCheckpoint = testCheckpoint;

// Cancel checkpoint
function cancelCheckpoint() {
    hideCheckpointSection();
    currentAccountID = null;
}

// Hide checkpoint section
function hideCheckpointSection() {
    const checkpointSection = document.getElementById('checkpointSection');
    const checkpointCode = document.getElementById('checkpointCode');
    const submitBtn = document.querySelector('button[onclick="solveCheckpoint()"]');

    // Clean up timer if it exists
    const timerElement = document.getElementById('expirationTimer');
    if (timerElement) {
        timerElement.remove();
    }

    // Reset form
    checkpointSection.style.display = 'none';
    checkpointCode.value = '';
    checkpointCode.disabled = false;
    submitBtn.disabled = false;

    // Reset checkpoint alert to default
    const checkpointAlert = document.getElementById('checkpointAlert');
    checkpointAlert.innerHTML = '<strong>Checkpoint Required:</strong> <span id="checkpointType"></span>';
}

// Disconnect LinkedIn account
async function disconnectLinkedIn(accountId) {
    if (!confirm('Are you sure you want to disconnect your LinkedIn account?')) {
        return;
    }

    // Find the specific disconnect button for this account
    const disconnectBtn = document.querySelector(`button[onclick="disconnectLinkedIn('${accountId}')"]`);
    let originalText = '';
    if (disconnectBtn) {
        originalText = disconnectBtn.innerHTML;
        disconnectBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Disconnecting...';
        disconnectBtn.disabled = true;
    }

    try {
        // Disconnect the LinkedIn account using the provided account ID
        const response = await fetch('/api/v1/accounts/linkedin', {
            method: 'DELETE',
            headers: getAuthHeaders(),
            body: JSON.stringify({
                account_id: accountId
            })
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
    } finally {
        // Restore button state
        if (disconnectBtn && originalText) {
            disconnectBtn.innerHTML = originalText;
            disconnectBtn.disabled = false;
        }
    }
}

// Resume checkpoint for a connecting account
async function resumeCheckpoint(accountId, checkpointType) {
    currentAccountID = accountId;

    // Show checkpoint section with the stored checkpoint type
    showCheckpointSection({
        type: checkpointType
    }, null);

    showAlert(`Resuming ${checkpointType} checkpoint for account ${accountId}`, 'info');
}

// Cancel connection for a connecting account
async function cancelConnection(accountId) {
    if (!confirm('Are you sure you want to cancel this connection?')) {
        return;
    }

    // Find the specific cancel button for this account
    const cancelBtn = document.querySelector(`button[onclick="cancelConnection('${accountId}')"]`);
    let originalText = '';
    if (cancelBtn) {
        originalText = cancelBtn.innerHTML;
        cancelBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Cancelling...';
        cancelBtn.disabled = true;
    }

    try {
        const response = await fetch('/api/v1/accounts/linkedin', {
            method: 'DELETE',
            headers: getAuthHeaders(),
            body: JSON.stringify({
                account_id: accountId
            })
        });

        const data = await response.json();

        if (response.ok) {
            showAlert('Connection cancelled successfully!', 'success');
            loadUserAccounts();
        } else {
            showAlert(data.error || 'Failed to cancel connection', 'danger');
        }
    } catch (error) {
        showAlert('Network error. Please try again.', 'danger');
    } finally {
        // Restore button state
        if (cancelBtn && originalText) {
            cancelBtn.innerHTML = originalText;
            cancelBtn.disabled = false;
        }
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
