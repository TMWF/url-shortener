package model

type URLBatchRequestDto struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type URLBatchResponseDto struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
