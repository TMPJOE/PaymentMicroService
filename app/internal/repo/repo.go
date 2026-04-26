// Package repo implements the data access layer of the application.
// It handles all database queries, transactions, and data mapping,
// providing a clean interface for the service layer to interact with PostgreSQL.
package repo

import ()

type ServiceRepository interface {
	DbPing() error
}

//REMEMBER TRANSACTION CODE LOGIC
