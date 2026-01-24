package main

import (
	"log"
	"net/http"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/handler"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/middleware"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.InitialiseConfigs()

	err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		logger.GetLogger().Fatal("Error occured while initialising logger")
	}
	defer logger.GetLogger().Sync()

	router := createRouter(cfg)
	logger.GetLogger().Info("Starting server on port " + cfg.ServerHost)

	log.Fatal(http.ListenAndServe(cfg.ServerHost, router))
}

func createRouter(config *config.Config) http.Handler {

	storage := repository.GetStorage(config)
	// urlStorage := repository.GetStorage(config)
	urlService := service.NewURLService(storage, config)
	urlHandler := handler.NewURLHandler(urlService)

	router := chi.NewRouter()
	router.Use(middleware.RequestLoggerMiddleware(logger.GetLogger()))
	router.Use(middleware.GzipMiddleware())
	router.Post(`/`, urlHandler.ShortenURL)
	router.Post(`/api/shorten`, urlHandler.ShortenURLAPI)
	router.Post(`/api/shorten/batch`, urlHandler.ShortenURLBatch)
	router.Get(`/{id}`, urlHandler.GetOriginalURL)

	if dbPinger, ok := storage.(repository.DBPinger); ok {
		pingService := service.NewPingDBService(dbPinger)
		pingHandler := handler.NewPingHandler(pingService)
		router.Get(`/ping`, pingHandler.PingDB)
	}

	return router
}
