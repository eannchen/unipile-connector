package repository

import "errors"

// Repositories is a collection of repositories
type Repositories struct {
	Tx      TxRepository
	User    UserRepository
	Account AccountRepository
}

// ErrDuplicateKey is returned when a duplicate key is encountered
var ErrDuplicateKey = errors.New("duplicate key")
