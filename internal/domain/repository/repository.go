package repository

import "errors"

// Repositories is a collection of repositories
type Repositories struct {
	Tx      TxRepository
	User    UserRepository
	Account AccountRepository
}

// ErrRecordNotFound is returned when a record is not found
var ErrRecordNotFound = errors.New("record not found")

// ErrDuplicateKey is returned when a duplicate key is encountered
var ErrDuplicateKey = errors.New("duplicate key")
