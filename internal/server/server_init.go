package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/config"
	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/handlers"
	"github.com/Piccadilly98/incidents_service/internal/health"
	"github.com/Piccadilly98/incidents_service/internal/middleware"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/repository/cache"
	"github.com/Piccadilly98/incidents_service/internal/repository/db"
	"github.com/Piccadilly98/incidents_service/internal/repository/queue"
	"github.com/Piccadilly98/incidents_service/internal/service"
	"github.com/Piccadilly98/incidents_service/internal/webhook_manager"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func ServerStart() (chan error, error) {
	cfg, err := config.NewConfig(true)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:        cfg.RedisAddr,
		Password:    cfg.RedisPassword,
		DialTimeout: 2 * time.Second,
	})

	queue, err := queue.NewRedisQueue(redisClient, context.Background(), 10)
	if err != nil {
		return nil, err
	}
	wm, err := webhook_manager.NewWebhookManager(cfg, queue, cfg.WebhookMaxReTry, true, context.Background())
	if err != nil {
		return nil, err
	}
	cache, err := cache.NewRedisCache(redisClient, context.Background(), cfg.RedisTTL)
	if err != nil {
		return nil, err
	}

	db, err := db.NewDB(cfg.ConnectionStr)
	if err != nil {
		return nil, err
	}
	healthChecker := health.NewHealthChecker([]health.Checks{db, queue, cache})
	service := service.NewService(db, cache, cfg, wm)
	r := chi.NewRouter()
	ew := error_worker.NewErrorWorker(cfg.LoggingUserError)
	regHandler, err := handlers.NewRegistrationHandler(service, ew)
	if err != nil {
		return nil, err
	}
	get, err := handlers.NewGetHandler(service, ew)
	if err != nil {
		return nil, err
	}
	healthHandler := handlers.NewHealthHandler(healthChecker, ew)
	updateHandler, err := handlers.NewUpdateHandler(service, ew)
	if err != nil {
		return nil, err
	}
	del, err := handlers.NewDeactivateHandler(service, ew)
	if err != nil {
		return nil, err
	}
	pagination, err := handlers.NewPaginationHandler(service, ew)
	if err != nil {
		return nil, err
	}
	lockCheck, err := handlers.NewLocationCheckHandler(service, ew)
	if err != nil {
		return nil, err
	}
	staticHandler, err := handlers.NewStatisticHandler(service, ew)
	if err != nil {
		return nil, err
	}
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY not set in .env")
	}
	fmt.Printf("\n\n\nAPI Key (для админских эндпоинтов): %-36s \n", apiKey)
	fmt.Printf("Используй в заголовке: X-API-Key: %s \n\n\n", apiKey)
	mid := middleware.CheckMiddleware(apiKey)
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/location/check", lockCheck.Handler)
		r.Get("/system/health", healthHandler.Handler)
		r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
			v := dto.ResultWebhookRequestDTO{}
			err := json.NewDecoder(r.Body).Decode(&v)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			log.Println(v)
			w.WriteHeader(http.StatusOK)
		})
		r.Group(func(r chi.Router) {
			r.Use(mid)
			r.Get("/incidents/stats", staticHandler.Handler)
			r.Delete("/incidents/{id}", del.Handler)
			r.Post("/incidents", regHandler.Handler)
			r.Put("/incidents/{id}", updateHandler.Handler)
			r.Get("/incidents/{id}", get.Handler)
			r.Get("/incidents", pagination.Handler)
		})
	})
	errCh := make(chan error)
	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.ServerAddr, cfg.ServerPort)
		log.Printf("Starting HTTP server on %s", addr)
		if err := http.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()
	return errCh, nil
}
