package model

// ShortenURLRequest описывает тело запроса на сокращение URL.
//
// Используется в API-методе создания короткой ссылки. Поле URL содержит исходный
// адрес, который необходимо сократить.
type ShortenURLRequest struct {
	// URL содержит исходный URL, который необходимо сократить.
	URL string `json:"url"`
}

// ShortenURLResponse описывает тело ответа на запрос сокращения URL.
//
// Возвращается API-методом создания короткой ссылки и содержит итоговый
// сокращённый URL.
type ShortenURLResponse struct {
	// ShortenedURL содержит сокращённый URL, созданный для исходного адреса.
	ShortenedURL string `json:"result"`
}
