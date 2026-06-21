package model

// URLBatchRequestDto описывает элемент запроса на пакетное сокращение URL.
//
// Структура используется в API пакетного сокращения ссылок. Каждый элемент
// содержит корреляционный идентификатор, позволяющий сопоставить входной URL
// с соответствующим результатом, и исходный URL, который необходимо сократить.
type URLBatchRequestDto struct {
	// CorrelationID содержит идентификатор, связывающий элемент запроса
	// с соответствующим элементом ответа.
	CorrelationID string `json:"correlation_id"`

	// OriginalURL содержит исходный URL, который необходимо сократить.
	OriginalURL string `json:"original_url"`
}

// URLBatchResponseDto описывает элемент ответа на пакетное сокращение URL.
//
// Структура возвращается API после обработки пакетного запроса. Каждый элемент
// содержит корреляционный идентификатор из исходного запроса и созданный короткий URL.
type URLBatchResponseDto struct {
	// CorrelationID содержит идентификатор, связывающий элемент ответа
	// с соответствующим элементом запроса.
	CorrelationID string `json:"correlation_id"`

	// ShortURL содержит сокращённый URL, созданный для исходного URL.
	ShortURL string `json:"short_url"`
}
