package repository

import (
	"context"

	"github.com/TMWF/url-shortener/internal/model"
)

type URLStorage interface {
	SaveURL(ctx context.Context, url string) (string, error)
	GetURL(ctx context.Context, id string) (string, error)
	SaveBatchURL(ctx context.Context, urlBatch []model.URLBatchRequestDto) ([]string, error)
	GetUsersURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error)
}
