package model

type GetUserURLsResponseModel struct {
	ShortUrl    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
