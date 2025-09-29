# LinkedIn Connector with Unipile

A full-stack web application that allows users to connect their LinkedIn accounts using Unipile's native authentication. The application supports both credential-based and cookie-based authentication methods with checkpoint handling.

### Completed

- JWT-based User Authentication
- LinkedIn Connection
  - Username/password authentication
  - Cookie-based authentication (li_at token)
- 2FA/OTP Checkpoint Handling
- Clean Architecture
  - e.g., Repository and Use Case Patterns
  - Dependency Injection


### Incomplete

- Checkpoint Handling
  - IN_APP_VALIDATION
    - Webhook Integration
    - WebSocket Support (for Real-time frontend updates)
  - CAPTCHA
  - PHONE_REGISTER
- Corn Job (Remove accounts where checkpoints are expired)
- Error Recovery
- Rate Limiting
- Security Enhancements
- Code structure Enhancements
- Testing
- Monitoring