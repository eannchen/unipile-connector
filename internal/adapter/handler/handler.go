package handler

// Handlers handles all requests
type Handlers struct {
	AuthHandler    AuthHandler
	AccountHandler AccountHandler
}

// NewHandlers creates a new handlers
func NewHandlers(authHandler AuthHandler, accountHandler AccountHandler) *Handlers {
	return &Handlers{AuthHandler: authHandler, AccountHandler: accountHandler}
}
