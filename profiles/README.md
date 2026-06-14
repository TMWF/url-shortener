### Выводы по инкременту 17

Замена использования encoding/json на "github.com/goccy/go-json" и прямая запись в http.ResponseWriter через json.NewEncoder().Encode() аозволило немного увеличить скорость работы хендлеров и уменьшить потребление памяти.