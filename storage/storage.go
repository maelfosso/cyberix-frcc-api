package storage

import "database/sql"

type Storage interface {
	Querier
	QuerierTx
}

type SQLStorage struct {
	db *sql.DB
	*Queries
}

func NewStorage(db *sql.DB) Storage {
	return &SQLStorage{
		db:      db,
		Queries: NewQueries(db),
	}
}
