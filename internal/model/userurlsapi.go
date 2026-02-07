package model

type GetUserURLsResponseModel struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
