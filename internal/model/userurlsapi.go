package model

// GetUserURLsResponseModel описывает элемент ответа со списком URL пользователя.
//
// Структура используется при возврате ссылок, созданных текущим пользователем.
// Содержит сокращённый URL и соответствующий ему исходный URL.
type GetUserURLsResponseModel struct {
	// ShortURL содержит сокращённый URL, созданный пользователем.
	ShortURL string `json:"short_url"`

	// OriginalURL содержит исходный URL, соответствующий сокращённой ссылке.
	OriginalURL string `json:"original_url"`
}
