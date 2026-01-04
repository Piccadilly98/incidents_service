package service

import (
	"log"
	"os"

	"github.com/Piccadilly98/incidents_service/internal/config"
	"github.com/Piccadilly98/incidents_service/internal/repository"
)

const (
	StatusActive   = "active"
	StatusResolved = "resolved"
)

type Service struct {
	cache            repository.CacheReposytory
	db               repository.DbReposytory
	config           *config.Config
	changeLogger     *log.Logger
	dbCriticalLogger *log.Logger
}

func NewService(db repository.DbReposytory, cache repository.CacheReposytory, config *config.Config) *Service {
	return &Service{
		db:               db,
		cache:            cache,
		config:           config,
		changeLogger:     log.New(os.Stdout, "[UPDATE SUBS INFO] ", log.Ldate|log.Ltime),
		dbCriticalLogger: log.New(os.Stderr, "[DB PING ERROR] ", log.Ldate|log.Ltime),
	}
}
