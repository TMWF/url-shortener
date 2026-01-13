package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/database"
	"github.com/TMWF/url-shortener/internal/handler"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/middleware"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, dbConfig := config.InitialiseConfigs()

	db, err := database.NewDB(dbConfig.DatabaseDSN)
	if err != nil {
		logger.GetLogger().Fatal("Error occured when creating DB connection")
	}

	err = logger.Initialize(cfg.LogLevel)
	if err != nil {
		logger.GetLogger().Fatal("Error occured while initialising logger")
	}
	defer logger.GetLogger().Sync()

	router := createRouter(cfg, db)
	logger.GetLogger().Info("Starting server on port " + cfg.ServerHost)

	log.Fatal(http.ListenAndServe(cfg.ServerHost, router))
}

func createRouter(config *config.Config, db *sql.DB) http.Handler {

	dbStorage := repository.NewDBStorage(db)
	pingService := service.NewPingDBService(dbStorage)
	pingHandler := handler.NewPingHandler(pingService)

	urlStorage := repository.NewFileStorage(config)
	urlService := service.NewURLService(urlStorage, config)
	urlHandler := handler.NewURLHandler(urlService)

	router := chi.NewRouter()
	router.Use(middleware.RequestLoggerMiddleware(logger.GetLogger()))
	router.Use(middleware.GzipMiddleware())
	router.Post(`/`, urlHandler.ShortenURL)
	router.Post(`/api/shorten`, urlHandler.ShortenURLAPI)
	router.Get(`/{id}`, urlHandler.GetOriginalURL)
	router.Get(`/ping`, pingHandler.PingDB)
	return router
}
