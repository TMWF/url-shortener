package model

// URLModel описывает URL, сохранённый в системе.
//
// Структура представляет модель ссылки и содержит внутренний идентификатор,
// сокращённый URL и исходный URL. Используется для хранения, передачи и
// сериализации данных о ссылке.
type URLModel struct {
	// UUID содержит уникальный идентификатор записи URL в системе.
	UUID string `json:"uuid"`

	// ShortURL содержит сокращённый URL.
	ShortURL string `json:"short_url"`

	// OriginalURL содержит исходный URL.
	OriginalURL string `json:"original_url"`
}
