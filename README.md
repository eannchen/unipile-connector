# LinkedIn Connector with Unipile

![CI](https://github.com/eannchen/unipile-connector/workflows/CI/badge.svg)

A full-stack web demonstration application enables users to connect their LinkedIn accounts with Unipile’s native authentication mechanism. The application supports both credential-based and cookie-based authentication, with proper checkpoint handling.

> The Unipile API trial on the demonstration site ([unipile-connector.onrender.com](https://unipile-connector.onrender.com)) will conclude on October 4, 2025.

### Completed

- JWT-based User Authentication
- LinkedIn Connection
  - Username/password authentication
  - Cookie-based authentication (`li_at` token)
- Checkpoint Handling
  - `2FA/OTP`
  - `PHONE_REGISTER`
- Migrations
- Error Handling
- Security Enhancements
  - CORS
  - Rate Limiting (memory store for deployment simplicity)
  - JWT token blacklisting (memory store for deployment simplicity)
- Clean Architecture
- Testing
- GitHub Actions CI for auto testing
- Backend Code Cleanup

### Incomplete

- Checkpoint Handling
  - `IN_APP_VALIDATION`
    - Webhook Integration
    - WebSocket Support (for Real-time frontend updates)
  - `CAPTCHA`
- UI/UX Enhancement for multiple accounts verification
- Frontend Code Cleanup

### Flow for `IN_APP_VALIDATION` Checkpoint

#### Initiate Connection
1. User initiates LinkedIn connection → Backend calls Unipile API
2. Unipile returns `IN_APP_VALIDATION` checkpoint → Webhook automatically registered
3. Frontend shows app validation UI → WebSocket connection established

#### Validate
1. User confirms in LinkedIn app → Unipile sends webhook to backend
2. Backend processes webhook → Broadcasts via WebSocket to frontend
3. Frontend receives update → Shows success message and refreshes account list

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   LinkedIn App  │    │   Unipile API    │    │     Backend     │
│                 │    │                  │    │                 │
│ User confirms   │───▶│   Status: "OK"   │───▶│ Webhook Handler │
│ connection      │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │  WebSocket Hub  │
                                               │                 │
                                               │  Broadcast to   │
                                               │ connected users │
                                               └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │    Frontend     │
                                               │                 │
                                               │    Real-time    │
                                               │  status update  │
                                               └─────────────────┘
```
