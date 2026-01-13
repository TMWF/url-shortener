package service

import "github.com/TMWF/url-shortener/internal/repository"

type PingDBService interface {
	PingDB() error
}

type pingDBServiceImpl struct {
	storage repository.DBStorage
}

func NewPingDBService(storage repository.DBStorage) *pingDBServiceImpl {
	return &pingDBServiceImpl{storage: storage}
}

func (ps *pingDBServiceImpl) PingDB() error {
	return ps.storage.PingDB()
}
