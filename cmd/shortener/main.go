package main

import (
	"log"
	"net/http"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/handler"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Config{}
	cfg.ParseFlags()
	router := createRouter(cfg)

	log.Fatal(http.ListenAndServe(cfg.ServerHost, router))
}

func createRouter(config config.Config) http.Handler {
	urlStorage := repository.NewMemStorage()
	urlService := service.NewURLService(urlStorage, &config)
	urlHandler := handler.NewURLHandler(urlService)

	router := chi.NewRouter()
	router.Post(`/`, urlHandler.ShortenURL)
	router.Get(`/{id}`, urlHandler.GetOriginalURL)
	return router
}
