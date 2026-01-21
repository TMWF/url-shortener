package model

type URLBatchRequestDto struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type URLBatchResponseDto struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
