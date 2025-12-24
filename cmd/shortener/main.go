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
	config.ParseFlags()
	urlStorage := repository.NewMemStorage()
	urlService := service.NewURLService(urlStorage)
	urlHandler := handler.NewURLHandler(urlService)

	router := chi.NewRouter()
	router.Post(`/`, urlHandler.ShortenURL)
	router.Get(`/{id}`, urlHandler.GetOriginalURL)

	log.Fatal(http.ListenAndServe(config.ServerHost, router))
}
