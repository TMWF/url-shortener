package repository

import "database/sql"

type DBStorage interface {
	PingDB() error
}

type dbStorageImpl struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *dbStorageImpl {
	return &dbStorageImpl{db: db}
}

func (dbs *dbStorageImpl) PingDB() error {
	return dbs.db.Ping()
}
