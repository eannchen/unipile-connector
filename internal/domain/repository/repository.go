package repository

// Repositories is a collection of repositories
type Repositories struct {
	Tx      TxRepository
	User    UserRepository
	Account AccountRepository
}
