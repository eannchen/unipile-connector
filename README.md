# LinkedIn Connector with Unipile

A full-stack web application that allows users to connect their LinkedIn accounts using Unipile's native authentication. The application supports both credential-based and cookie-based authentication methods with checkpoint handling.

### Completed

- JWT-based User Authentication
- LinkedIn Connection
  - Username/password authentication
  - Cookie-based authentication (`li_at` token)
- Checkpoint Handling
  - `2FA/OTP`
- Clean Architecture
  - e.g., Repository and Use Case Patterns
  - Dependency Injection

### Incomplete

- Checkpoint Handling
  - `IN_APP_VALIDATION`
    - Webhook Integration
    - WebSocket Support (for Real-time frontend updates)
  - `CAPTCHA`
  - `PHONE_REGISTER`
- Corn Job (Remove accounts where checkpoints are expired)
- Security Enhancements
  - Rate Limiting
  - CORS
- Better Error Handling for UX
- Code Cleanup and Refactoring
- Testing


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
