package main

import (
	"net/http"

	"github.com/TMWF/url-shortener/internal/handler"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/service"
)

func main() {
	urlStorage := repository.NewMemStorage()
	urlService := service.NewURLService(urlStorage)
	urlHandler := handler.NewURLHandler(*urlService)
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, urlHandler.ShortenURL)
	mux.HandleFunc(`/{id}`, urlHandler.GetOriginalURL)

	err := http.ListenAndServe(`localhost:8080`, mux)
	if err != nil {
		panic(err)
	}
}
