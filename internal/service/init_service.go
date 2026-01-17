package service

import (
	"log"
	"os"

	"github.com/Piccadilly98/incidents_service/internal/config"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/repository"
)

type WebhookSender interface {
	AddToQueue(result dto.LocationCheckResponse, url, method string)
	Stop()
}

type Service struct {
	cache            repository.CacheReposytory
	db               repository.DbReposytory
	config           *config.Config
	changeLogger     *log.Logger
	dbCriticalLogger *log.Logger
	cacheLogger      *log.Logger
	wm               WebhookSender
}

func NewService(db repository.DbReposytory, cache repository.CacheReposytory, config *config.Config, wm WebhookSender) *Service {
	return &Service{
		db:               db,
		cache:            cache,
		config:           config,
		changeLogger:     log.New(os.Stdout, "[UPDATE INCIDENT INFO] ", log.Ldate|log.Ltime),
		dbCriticalLogger: log.New(os.Stderr, "[DB PING ERROR] ", log.Ldate|log.Ltime),
		cacheLogger:      log.New(os.Stderr, "[CACHE ERROR]", log.Ldate|log.Ltime),
		wm:               wm,
	}
}
